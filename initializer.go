package pathvars

var initializers []func()
var initialized bool

func RegisterInitializer(initializer func()) {
	initializers = append(initializers, initializer)
}

func initialize() {
	if initialized {
		goto end
	}
	for _, initializer := range initializers {
		initializer()
	}
	initialized = true
end:
	return
}
