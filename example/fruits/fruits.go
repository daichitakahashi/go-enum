package fruits

import "github.com/daichitakahashi/go-enum"

//go:generate go run github.com/daichitakahashi/go-enum/cmd/enumgen@latest --visitor-impl "*"

type (
	Fruits interface{}

	Apple  enum.MemberOf[Fruits]
	Orange enum.MemberOf[Fruits]
	Grape  enum.MemberOf[Fruits]
)
