package enum

// MemberOf marks membership of EnumIdent on struct.
type MemberOf[EnumIdent any] struct{}

// VisitorReturns specifies return type of visitor on enum identifier.
type VisitorReturns[Return any] interface{}
