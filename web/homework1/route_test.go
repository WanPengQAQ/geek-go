package web

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"reflect"
	"testing"
)

// 测试输入path有问题的case
func TestMyTestOfIllegalPath(t *testing.T) {
	mockHandler := func(ctx *Context) {}

	r := newRouter()

	// 空字符串
	assert.PanicsWithValue(t, "web: 路由是空字符串", func() {
		r.addRoute(http.MethodGet, "", mockHandler)
	})

	// 前导没有 /
	assert.PanicsWithValue(t, "web: 路由必须以 / 开头", func() {
		r.addRoute(http.MethodGet, "a/b/c", mockHandler)
	})

	// 后缀有 /
	assert.PanicsWithValue(t, "web: 路由不能以 / 结尾", func() {
		r.addRoute(http.MethodGet, "/a/b/c/", mockHandler)
	})

	// 多个 /
	assert.PanicsWithValue(t, "web: 非法路由。不允许使用 //a/b, /a//b 之类的路由, [/a//b]", func() {
		r.addRoute(http.MethodGet, "/a//b", mockHandler)
	})
	assert.PanicsWithValue(t, "web: 非法路由。不允许使用 //a/b, /a//b 之类的路由, [//a/b]", func() {
		r.addRoute(http.MethodGet, "//a/b", mockHandler)
	})

	// 普通节点重复注册
	r.addRoute(http.MethodGet, "/a/b/c", mockHandler)
	assert.PanicsWithValue(t, "web: 路由冲突[/a/b/c]", func() {
		r.addRoute(http.MethodGet, "/a/b/c", mockHandler)
	})

	// 根节点重复注册
	r.addRoute(http.MethodGet, "/", mockHandler)
	assert.PanicsWithValue(t, "web: 路由冲突[/]", func() {
		r.addRoute(http.MethodGet, "/", mockHandler)
	})

	// 同时注册通配符路由，参数路由，正则路由
	r = newRouter()
	assert.PanicsWithValue(t, "web: 非法路由，已有通配符路由。不允许同时注册通配符路由和参数路由 [:id]", func() {
		r.addRoute(http.MethodGet, "/a/*", mockHandler)
		r.addRoute(http.MethodGet, "/a/:id", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "web: 非法路由，已有通配符路由。不允许同时注册通配符路由和正则路由 [:id(.*)]", func() {
		r.addRoute(http.MethodGet, "/a/b/*", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/:id(.*)", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "web: 非法路由，已有通配符路由。不允许同时注册通配符路由和参数路由 [:id]", func() {
		r.addRoute(http.MethodGet, "/*", mockHandler)
		r.addRoute(http.MethodGet, "/:id", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "web: 非法路由，已有路径参数路由。不允许同时注册通配符路由和参数路由 [*]", func() {
		r.addRoute(http.MethodGet, "/a/b/:id", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/*", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "web: 非法路由，已有路径参数路由。不允许同时注册通配符路由和参数路由 [*]", func() {
		r.addRoute(http.MethodGet, "/:id", mockHandler)
		r.addRoute(http.MethodGet, "/*", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "web: 非法路由，已有路径参数路由。不允许同时注册正则路由和参数路由 [:id(.*)]", func() {
		r.addRoute(http.MethodGet, "/a/b/:id", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/:id(.*)", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "web: 非法路由，已有正则路由。不允许同时注册通配符路由和正则路由 [*]", func() {
		r.addRoute(http.MethodGet, "/a/b/:id(.*)", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/*", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "web: 非法路由，已有正则路由。不允许同时注册正则路由和参数路由 [:id]", func() {
		r.addRoute(http.MethodGet, "/a/b/:id(.*)", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/:id", mockHandler)
	})
	// 参数冲突
	r = newRouter()
	assert.PanicsWithValue(t, "web: 路由冲突，参数路由冲突，已有 :id，新注册 :name", func() {
		r.addRoute(http.MethodGet, "/a/b/c/:id", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/c/:name", mockHandler)
	})
	// 正则冲突
	r = newRouter()
	assert.PanicsWithValue(t, "web: 路由冲突，正则路由冲突，已有 :id(.*)，新注册 :name([0-9])", func() {
		r.addRoute(http.MethodGet, "/a/b/*/123/:id(.*)", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/*/123/:name([0-9])", mockHandler)
	})

	r = newRouter()
	assert.NotPanics(t, func() {
		r.addRoute(http.MethodGet, "/a/b/:id", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/:id/123", mockHandler)
	})

	r = newRouter()
	assert.NotPanics(t, func() {
		r.addRoute(http.MethodGet, "/a/b/:id(.*)", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/:id(.*)/123", mockHandler)
	})

}

func TestMyTestOfStaticRoute(t *testing.T) {
	testRoutes := []struct {
		method []string
		path   []string
		should router
	}{
		{
			method: []string{http.MethodGet, http.MethodPost, http.MethodGet},
			path:   []string{"/", "/", "/*"},
			should: router{trees: map[string]*node{
				http.MethodGet: &node{
					typ:  nodeTypeStatic,
					path: "/",
					starChild: &node{
						typ:  nodeTypeAny,
						path: "*",
					},
				},
				http.MethodPost: &node{
					path: "/",
				},
			}},
		},
		{
			method: []string{http.MethodGet, http.MethodGet},
			path:   []string{"/a", "/a/*"},
			should: router{trees: map[string]*node{
				http.MethodGet: &node{
					path: "/",
					children: map[string]*node{
						"a": &node{
							path: "a",
							starChild: &node{
								path: "*",
							},
						},
					},
				},
			}},
		},
		{
			method: []string{http.MethodGet, http.MethodGet, http.MethodGet, http.MethodGet, http.MethodGet},
			path:   []string{"/a", "/b", "/a/c", "/a/e", "/*/c"},
			should: router{trees: map[string]*node{
				http.MethodGet: &node{
					path: "/",
					children: map[string]*node{
						"a": &node{
							path: "a",
							children: map[string]*node{
								"c": &node{
									path: "c",
								},
								"e": &node{
									path: "e",
								},
							},
						},
						"b": &node{
							path: "b",
						},
					},
					starChild: &node{
						typ:  nodeTypeAny,
						path: "*",
						children: map[string]*node{
							"c": &node{
								path: "c",
							},
						},
					},
				},
			}},
		},
	}

	for _, u := range testRoutes {
		r := newRouter()
		for i := 0; i < len(u.method); i++ {
			r.addRoute(u.method[i], u.path[i], nil)
		}
		s, ok := r.equal(u.should)
		assert.Equal(t, ok, true, s)
	}
}

func TestMyAddRoute(t *testing.T) {
	addActions := []struct {
		method string
		path   string
	}{
		// 静态路由
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user/home",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodPost,
			path:   "/login",
		},
		// 通配符测试用例
		{
			method: http.MethodGet,
			path:   "/order/*",
		},
		{
			method: http.MethodGet,
			path:   "/*",
		},
		{
			method: http.MethodGet,
			path:   "/*/*",
		},
		{
			method: http.MethodGet,
			path:   "/*/abc",
		},
		{
			method: http.MethodGet,
			path:   "/*/abc/*",
		},
		// 参数查找测试用例
		{
			method: http.MethodGet,
			path:   "/param/:id",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id/detail",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id/*",
		},
		// 正则路由
		{
			method: http.MethodDelete,
			path:   "/reg/:id(.*)",
		},
		{
			method: http.MethodDelete,
			path:   "/:name(^.+$)/abc",
		},
	}
	mokeHandle := func(ctx *Context) {}
	should := &router{trees: map[string]*node{
		http.MethodGet: &node{
			typ:     nodeTypeStatic,
			path:    "/",
			handler: mokeHandle,
			children: map[string]*node{
				"user": &node{
					typ:     nodeTypeStatic,
					path:    "user",
					handler: mokeHandle,
					children: map[string]*node{
						"home": &node{
							typ:     nodeTypeStatic,
							path:    "home",
							handler: mokeHandle,
						},
					},
				},
				"order": &node{
					typ:  nodeTypeStatic,
					path: "order",
					children: map[string]*node{
						"detail": &node{
							typ:     nodeTypeStatic,
							path:    "detail",
							handler: mokeHandle,
						},
					},
					starChild: &node{
						typ:     nodeTypeAny,
						path:    "*",
						handler: mokeHandle,
					},
				},
				"param": &node{
					typ:  nodeTypeStatic,
					path: "param",
					paramChild: &node{
						typ:       nodeTypeParam,
						path:      ":id",
						handler:   mokeHandle,
						paramName: "id",
						children: map[string]*node{
							"detail": &node{
								typ:     nodeTypeStatic,
								path:    "detai",
								handler: mokeHandle,
							},
						},
					},
				},
			},
			starChild: &node{
				typ:     nodeTypeAny,
				path:    "*",
				handler: mokeHandle,
				starChild: &node{
					typ:     nodeTypeAny,
					path:    "*",
					handler: mokeHandle,
				},
				children: map[string]*node{
					"abc": &node{
						typ:     nodeTypeStatic,
						path:    "abc",
						handler: mokeHandle,
						starChild: &node{
							typ:     nodeTypeAny,
							path:    "*",
							handler: mokeHandle,
						},
					},
				},
			},
		},
		http.MethodPost: &node{
			typ:  nodeTypeStatic,
			path: "/",
			children: map[string]*node{
				"login": &node{
					typ:     nodeTypeStatic,
					path:    "login",
					handler: mokeHandle,
				},
				"order": &node{
					typ:  nodeTypeStatic,
					path: "order",
					children: map[string]*node{
						"create": &node{
							typ:     nodeTypeStatic,
							path:    "create",
							handler: mokeHandle,
						},
					},
				},
			},
		},
	}}

	r := newRouter()
	for _, act := range addActions {
		r.addRoute(act.method, act.path, mokeHandle)
	}

	errMsg, ok := should.equal(r)
	if !ok {
		assert.Equal(t, ok, true, errMsg)
	}
}

func Test_router_AddRoute(t *testing.T) {
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/user/home",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodPost,
			path:   "/login",
		},
		// 通配符测试用例
		{
			method: http.MethodGet,
			path:   "/order/*",
		},
		{
			method: http.MethodGet,
			path:   "/*",
		},
		{
			method: http.MethodGet,
			path:   "/*/*",
		},
		{
			method: http.MethodGet,
			path:   "/*/abc",
		},
		{
			method: http.MethodGet,
			path:   "/*/abc/*",
		},
		// 参数路由
		{
			method: http.MethodGet,
			path:   "/param/:id",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id/detail",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id/*",
		},
		// 正则路由
		{
			method: http.MethodDelete,
			path:   "/reg/:id(.*)",
		},
		{
			method: http.MethodDelete,
			path:   "/:name(^.+$)/abc",
		},
	}

	mockHandler := func(ctx *Context) {}
	r := newRouter()
	for _, tr := range testRoutes {
		r.addRoute(tr.method, tr.path, mockHandler)
	}

	wantRouter := &router{
		trees: map[string]*node{
			http.MethodGet: {
				path: "/",
				children: map[string]*node{
					"user": {
						path: "user",
						children: map[string]*node{
							"home": {path: "home", handler: mockHandler, typ: nodeTypeStatic},
						},
						handler: mockHandler,
						typ:     nodeTypeStatic,
					},
					"order": {
						path: "order",
						children: map[string]*node{
							"detail": {path: "detail", handler: mockHandler, typ: nodeTypeStatic},
						},
						starChild: &node{path: "*", handler: mockHandler, typ: nodeTypeAny},
						typ:       nodeTypeStatic,
					},
					"param": {
						path: "param",
						paramChild: &node{
							path:      ":id",
							paramName: "id",
							starChild: &node{
								path:    "*",
								handler: mockHandler,
								typ:     nodeTypeAny,
							},
							children: map[string]*node{"detail": {path: "detail", handler: mockHandler, typ: nodeTypeStatic}},
							handler:  mockHandler,
							typ:      nodeTypeParam,
						},
					},
				},
				starChild: &node{
					path: "*",
					children: map[string]*node{
						"abc": {
							path:      "abc",
							starChild: &node{path: "*", handler: mockHandler, typ: nodeTypeAny},
							handler:   mockHandler,
							typ:       nodeTypeStatic,
						},
					},
					starChild: &node{path: "*", handler: mockHandler, typ: nodeTypeAny},
					handler:   mockHandler,
					typ:       nodeTypeAny,
				},
				handler: mockHandler,
				typ:     nodeTypeStatic,
			},
			http.MethodPost: {
				path: "/",
				children: map[string]*node{
					"order": {path: "order", children: map[string]*node{
						"create": {path: "create", handler: mockHandler, typ: nodeTypeStatic},
					}},
					"login": {path: "login", handler: mockHandler, typ: nodeTypeStatic},
				},
				typ: nodeTypeStatic,
			},
			http.MethodDelete: {
				path: "/",
				children: map[string]*node{
					"reg": {
						path: "reg",
						typ:  nodeTypeStatic,
						regChild: &node{
							path:      ":id(.*)",
							paramName: "id",
							typ:       nodeTypeReg,
							handler:   mockHandler,
						},
					},
				},
				regChild: &node{
					path:      ":name(^.+$)",
					paramName: "name",
					typ:       nodeTypeReg,
					children: map[string]*node{
						"abc": {
							path:    "abc",
							handler: mockHandler,
						},
					},
				},
			},
		},
	}
	msg, ok := wantRouter.equal(r)
	assert.True(t, ok, msg)

	// 非法用例
	r = newRouter()

	// 空字符串
	assert.PanicsWithValue(t, "web: 路由是空字符串", func() {
		r.addRoute(http.MethodGet, "", mockHandler)
	})

	// 前导没有 /
	assert.PanicsWithValue(t, "web: 路由必须以 / 开头", func() {
		r.addRoute(http.MethodGet, "a/b/c", mockHandler)
	})

	// 后缀有 /
	assert.PanicsWithValue(t, "web: 路由不能以 / 结尾", func() {
		r.addRoute(http.MethodGet, "/a/b/c/", mockHandler)
	})

	// 根节点重复注册
	r.addRoute(http.MethodGet, "/", mockHandler)
	assert.PanicsWithValue(t, "web: 路由冲突[/]", func() {
		r.addRoute(http.MethodGet, "/", mockHandler)
	})
	// 普通节点重复注册
	r.addRoute(http.MethodGet, "/a/b/c", mockHandler)
	assert.PanicsWithValue(t, "web: 路由冲突[/a/b/c]", func() {
		r.addRoute(http.MethodGet, "/a/b/c", mockHandler)
	})

	// 多个 /
	assert.PanicsWithValue(t, "web: 非法路由。不允许使用 //a/b, /a//b 之类的路由, [/a//b]", func() {
		r.addRoute(http.MethodGet, "/a//b", mockHandler)
	})
	assert.PanicsWithValue(t, "web: 非法路由。不允许使用 //a/b, /a//b 之类的路由, [//a/b]", func() {
		r.addRoute(http.MethodGet, "//a/b", mockHandler)
	})

	// 同时注册通配符路由，参数路由，正则路由
	assert.PanicsWithValue(t, "web: 非法路由，已有通配符路由。不允许同时注册通配符路由和参数路由 [:id]", func() {
		r.addRoute(http.MethodGet, "/a/*", mockHandler)
		r.addRoute(http.MethodGet, "/a/:id", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "web: 非法路由，已有通配符路由。不允许同时注册通配符路由和正则路由 [:id(.*)]", func() {
		r.addRoute(http.MethodGet, "/a/b/*", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/:id(.*)", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "web: 非法路由，已有通配符路由。不允许同时注册通配符路由和参数路由 [:id]", func() {
		r.addRoute(http.MethodGet, "/*", mockHandler)
		r.addRoute(http.MethodGet, "/:id", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "web: 非法路由，已有路径参数路由。不允许同时注册通配符路由和参数路由 [*]", func() {
		r.addRoute(http.MethodGet, "/a/b/:id", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/*", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "web: 非法路由，已有路径参数路由。不允许同时注册通配符路由和参数路由 [*]", func() {
		r.addRoute(http.MethodGet, "/:id", mockHandler)
		r.addRoute(http.MethodGet, "/*", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "web: 非法路由，已有路径参数路由。不允许同时注册正则路由和参数路由 [:id(.*)]", func() {
		r.addRoute(http.MethodGet, "/a/b/:id", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/:id(.*)", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "web: 非法路由，已有正则路由。不允许同时注册通配符路由和正则路由 [*]", func() {
		r.addRoute(http.MethodGet, "/a/b/:id(.*)", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/*", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "web: 非法路由，已有正则路由。不允许同时注册正则路由和参数路由 [:id]", func() {
		r.addRoute(http.MethodGet, "/a/b/:id(.*)", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/:id", mockHandler)
	})
	// 参数冲突
	assert.PanicsWithValue(t, "web: 路由冲突，参数路由冲突，已有 :id，新注册 :name", func() {
		r.addRoute(http.MethodGet, "/a/b/c/:id", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/c/:name", mockHandler)
	})
}

func (r router) equal(y router) (string, bool) {
	for k, v := range r.trees {
		yv, ok := y.trees[k]
		if !ok {
			return fmt.Sprintf("目标 router 里面没有方法 %s 的路由树", k), false
		}
		str, ok := v.equal(yv)
		if !ok {
			return k + "-" + str, ok
		}
	}
	return "", true
}

func (n *node) equal(y *node) (string, bool) {
	if y == nil {
		return "目标节点为 nil", false
	}
	if n.path != y.path {
		return fmt.Sprintf("%s 节点 path 不相等 x %s, y %s", n.path, n.path, y.path), false
	}

	nhv := reflect.ValueOf(n.handler)
	yhv := reflect.ValueOf(y.handler)
	if nhv != yhv {
		return fmt.Sprintf("%s 节点 handler 不相等 x %s, y %s", n.path, nhv.Type().String(), yhv.Type().String()), false
	}

	if n.typ != y.typ {
		return fmt.Sprintf("%s 节点类型不相等 x %d, y %d", n.path, n.typ, y.typ), false
	}

	if n.paramName != y.paramName {
		return fmt.Sprintf("%s 节点参数名字不相等 x %s, y %s", n.path, n.paramName, y.paramName), false
	}

	if len(n.children) != len(y.children) {
		return fmt.Sprintf("%s 子节点长度不等", n.path), false
	}
	if len(n.children) == 0 {
		return "", true
	}

	if n.starChild != nil {
		str, ok := n.starChild.equal(y.starChild)
		if !ok {
			return fmt.Sprintf("%s 通配符节点不匹配 %s", n.path, str), false
		}
	}
	if n.paramChild != nil {
		str, ok := n.paramChild.equal(y.paramChild)
		if !ok {
			return fmt.Sprintf("%s 路径参数节点不匹配 %s", n.path, str), false
		}
	}

	if n.regChild != nil {
		str, ok := n.regChild.equal(y.regChild)
		if !ok {
			return fmt.Sprintf("%s 路径参数节点不匹配 %s", n.path, str), false
		}
	}

	for k, v := range n.children {
		yv, ok := y.children[k]
		if !ok {
			return fmt.Sprintf("%s 目标节点缺少子节点 %s", n.path, k), false
		}
		str, ok := v.equal(yv)
		if !ok {
			return n.path + "-" + str, ok
		}
	}
	return "", true
}

func Test_router_findRoute(t *testing.T) {
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodGet,
			path:   "/user/*/home",
		},
		{
			method: http.MethodPost,
			path:   "/order/*",
		},
		// 参数路由
		{
			method: http.MethodGet,
			path:   "/param/:id",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id/detail",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id/*",
		},

		// 正则
		{
			method: http.MethodDelete,
			path:   "/reg/:id(.*)",
		},
		{
			method: http.MethodDelete,
			path:   "/:id([0-9]+)/home",
		},
	}

	mockHandler := func(ctx *Context) {}

	testCases := []struct {
		name   string
		method string
		path   string
		found  bool
		mi     *matchInfo
	}{
		{
			name:   "method not found",
			method: http.MethodHead,
		},
		{
			name:   "path not found",
			method: http.MethodGet,
			path:   "/abc",
		},
		{
			name:   "root",
			method: http.MethodGet,
			path:   "/",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "/",
					handler: mockHandler,
				},
			},
		},
		{
			name:   "user",
			method: http.MethodGet,
			path:   "/user",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "user",
					handler: mockHandler,
				},
			},
		},
		{
			name:   "no handler",
			method: http.MethodPost,
			path:   "/order",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path: "order",
				},
			},
		},
		{
			name:   "two layer",
			method: http.MethodPost,
			path:   "/order/create",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "create",
					handler: mockHandler,
				},
			},
		},
		// 通配符匹配
		{
			// 命中/order/*
			name:   "star match",
			method: http.MethodPost,
			path:   "/order/delete",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "*",
					handler: mockHandler,
				},
			},
		},
		{
			// 命中通配符在中间的
			// /user/*/home
			name:   "star in middle",
			method: http.MethodGet,
			path:   "/user/Tom/home",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "home",
					handler: mockHandler,
				},
			},
		},
		{
			// 比 /order/* 多了一段
			name:   "overflow",
			method: http.MethodPost,
			path:   "/order/delete/123",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "*",
					handler: mockHandler,
				},
			},
		},
		// 参数匹配
		{
			// 命中 /param/:id
			name:   ":id",
			method: http.MethodGet,
			path:   "/param/123",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    ":id",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
		{
			// 命中 /param/:id/*
			name:   ":id*",
			method: http.MethodGet,
			path:   "/param/123/abc",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "*",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
		{
			// 命中 /param/:id/detail
			name:   ":id*",
			method: http.MethodGet,
			path:   "/param/123/detail",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "detail",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
		{
			// 命中 /reg/:id(.*)
			name:   ":id(.*)",
			method: http.MethodDelete,
			path:   "/reg/123",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    ":id(.*)",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
		{
			// 命中 /:id([0-9]+)/home
			name:   ":id([0-9]+)",
			method: http.MethodDelete,
			path:   "/123/home",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "home",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
		{
			// 未命中 /:id([0-9]+)/home
			name:   "not :id([0-9]+)",
			method: http.MethodDelete,
			path:   "/abc/home",
		},
	}

	r := newRouter()
	for _, tr := range testRoutes {
		r.addRoute(tr.method, tr.path, mockHandler)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mi, found := r.findRoute(tc.method, tc.path)
			assert.Equal(t, tc.found, found)
			if !found {
				return
			}
			assert.Equal(t, tc.mi.pathParams, mi.pathParams)
			assert.Equal(t, tc.mi.n.path, mi.n.path)
			n := mi.n
			wantVal := reflect.ValueOf(tc.mi.n.handler)
			nVal := reflect.ValueOf(n.handler)
			assert.Equal(t, wantVal, nVal)
		})
	}
}
