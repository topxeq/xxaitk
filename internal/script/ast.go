package script

type NodeType int

const (
	NodeProgram NodeType = iota
	NodeLetStmt
	NodeConstStmt
	NodeAssignStmt
	NodeReturnStmt
	NodeIfStmt
	NodeWhileStmt
	NodeForStmt
	NodeFnDecl
	NodeFnLiteral
	NodeCallExpr
	NodeIndexExpr
	NodeDotExpr
	NodeBinaryExpr
	NodeUnaryExpr
	NodeIntLit
	NodeFloatLit
	NodeStringLit
	NodeBoolLit
	NodeNilLit
	NodeListLit
	NodeMapLit
	NodeIdentExpr
	NodeBreakStmt
	NodeContinueStmt
	NodeBreakpointStmt
	NodeBlockStmt
)

type Node struct {
	Type     NodeType
	Line     int
	Children []*Node
	Token    Token
	Value    string
	IntVal   int64
	FloatVal float64
}

func (n *Node) addChild(child *Node) {
	n.Children = append(n.Children, child)
}
