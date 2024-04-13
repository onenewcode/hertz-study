# hertz
## service
### options
首先定义一个option,每个WithReadTimeout都会生成一个option函数，之后在循环调用
```go
type Option struct {
	F func(o *Options)
}
```
在生成engine时候调用apply，进行循环调用
```go
func (o *Options) Apply(opts []Option) {
	for _, op := range opts {
		op.F(o)
	}
}
```
