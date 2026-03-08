package newapp

type Params interface {
	Param(key string) string
	BoolParam(key string) bool
}
