package web

import (
	"regexp"
	"strings"
)

type router struct {
	// trees 是按照 HTTP 方法来组织的
	// 如 GET => *node
	trees map[string]*node
}

func newRouter() router {
	return router{
		trees: map[string]*node{},
	}
}

// checkPath 校验path的合法性
func (r *router) checkPath(path string) {
	if len(path) == 0 {
		panic("web: 路由是空字符串")
	}
	if !strings.HasPrefix(path, "/") {
		panic("web: 路由必须以 / 开头")
	}
	if strings.HasSuffix(path, "/") && len(path) != 1 {
		panic("web: 路由不能以 / 结尾")
	}

	if strings.HasPrefix(path, "//") {
		panic("web: 非法路由。不允许使用 //a/b, /a//b 之类的路由, [" + path + "]")
	}

	if strings.Contains(path, "//") {
		panic("web: 非法路由。不允许使用 //a/b, /a//b 之类的路由, [" + path + "]")
	}
}

// 将输入的path根据"/"切分，例如:
//
//	/a/b/c -> [/, a, b, c]
//	 / -> [/]
func (r *router) pathSegment(path string) []string {
	sp := strings.Split(path, "/")
	result := []string{"/"}
	for _, seg := range sp {
		if len(strings.TrimSpace(seg)) == 0 {
			continue
		}
		result = append(result, seg)
	}
	return result
}

// addRoute 注册路由。
// method 是 HTTP 方法
// - 已经注册了的路由，无法被覆盖。例如 /user/home 注册两次，会冲突
// - path 必须以 / 开始并且结尾不能有 /，中间也不允许有连续的 /
// - 不能在同一个位置注册不同的参数路由，例如 /user/:id 和 /user/:name 冲突
// - 不能在同一个位置同时注册通配符路由和参数路由，例如 /user/:id 和 /user/* 冲突
// - 同名路径参数，在路由匹配的时候，值会被覆盖。例如 /user/:id/abc/:id，那么 /user/123/abc/456 最终 id = 456
func (r *router) addRoute(method string, path string, handler HandleFunc) {
	r.checkPath(path)
	r.initRoot(method)

	currNode := r.trees[method]
	segs := r.pathSegment(path)
	for _, seg := range segs[1:] {
		currNode = currNode.childOrCreate(seg)
	}

	if currNode.handler != nil {
		panic("web: 路由冲突[" + path + "]")
	}
	currNode.handler = handler
}

func (r *router) initRoot(method string) {
	root := r.trees[method]
	if root == nil {
		r.trees[method] = newRootNode()
	}
}

// findRoute 查找对应的节点
// 注意，返回的 node 内部 HandleFunc 不为 nil 才算是注册了路由
func (r *router) findRoute(method string, path string) (*matchInfo, bool) {
	panic("implement me")
}

type nodeType int

const (
	// 虚拟节点
	nodeTypeFake = iota
	// 静态路由
	nodeTypeStatic
	// 正则路由
	nodeTypeReg
	// 路径参数路由
	nodeTypeParam
	// 通配符路由
	nodeTypeAny
)

// node 代表路由树的节点
// 路由树的匹配顺序是：
// 1. 静态完全匹配
// 2. 正则匹配，形式 :param_name(reg_expr)
// 3. 路径参数匹配：形式 :param_name
// 4. 通配符匹配：*
// 这是不回溯匹配
type node struct {
	typ nodeType

	path string
	// children 子节点
	// 子节点的 path => node
	children map[string]*node
	// handler 命中路由之后执行的逻辑
	handler HandleFunc

	// 通配符 * 表达的节点，任意匹配
	starChild *node

	paramChild *node
	// 正则路由和参数路由都会使用这个字段
	paramName string

	// 正则表达式
	regChild *node
	regExpr  *regexp.Regexp
}

func newRootNode() *node {
	return &node{
		typ:        0,
		path:       "/",
		children:   nil,
		handler:    nil,
		starChild:  nil,
		paramChild: nil,
		paramName:  "",
		regChild:   nil,
		regExpr:    nil,
	}
}

// child 返回子节点
// 第一个返回值 *node 是命中的节点
// 第二个返回值 bool 代表是否命中
func (n *node) childOf(path string) (*node, bool) {
	panic("implement me")
}

// childOrCreate 查找子节点，
// 首先会判断 path 是不是通配符路径
// 其次判断 path 是不是参数路径，即以 : 开头的路径
// 最后会从 children 里面查找，
// 如果没有找到，那么会创建一个新的节点，并且保存在 node 里面
func (n *node) childOrCreate(path string) *node {
	if n, ok := n.children[path]; ok {
		return n
	}

	if n.children == nil {
		n.children = make(map[string]*node)
	}

	n.children[path] = &node{
		typ:        0,
		path:       path,
		children:   nil,
		handler:    nil,
		starChild:  nil,
		paramChild: nil,
		paramName:  "",
		regChild:   nil,
		regExpr:    nil,
	}
	return n.children[path]
}

type matchInfo struct {
	n          *node
	pathParams map[string]string
}

func (m *matchInfo) addValue(key string, value string) {
	if m.pathParams == nil {
		// 大多数情况，参数路径只会有一段
		m.pathParams = map[string]string{key: value}
	}
	m.pathParams[key] = value
}
