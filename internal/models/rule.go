package models

type Rule struct {
	ID          int64
	RequestType string
	Condition   map[string]interface{}
	Action      string
	GradeID     int64
	Active      bool
}
