package rules_lib

func RegisterValueDecoder(
	decoderName string,
	decoderFunc func(string) (interface{}, error),
) {
	_, found := valueDecoders[decoderName]
	if found {
		panic("duplicate rule value decoder name: " + decoderName)
	}
	valueDecoders[decoderName] = decoderFunc
}

func RegisterRuleType(
	order int,
	typeName string,
	decoderName string,
	checker *func(value interface{}) bool,
) {
	_, found := ruleTypes[typeName]
	if found {
		panic("duplicate rule type: " + typeName)
	}
	decoderFunc, ok := valueDecoders[decoderName]
	if !ok {
		panic("invalid rule value decoder: " + decoderName)
	}
	ruleTypes[typeName] = &EventRuleType{
		Order:        order,
		Name:         typeName,
		ValueDecoder: decoderFunc,
		ValueChecker: checker,
	}
}
