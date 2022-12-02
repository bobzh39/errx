package errx

import "fmt"

type FieldContext interface {
	Append(...LogField)
}

type LogField struct {
	key   string
	value any
}

func Field(key string, value any) LogField {
	return LogField{key, value}
}

func (l LogField) String() string {
	return fmt.Sprintf("%s=%v", l.key, l.value)
}
