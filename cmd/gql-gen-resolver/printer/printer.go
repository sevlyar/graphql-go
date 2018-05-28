package printer

import (
	"bytes"
	"fmt"
	"github.com/sevlyar/graphql-go"
	"github.com/sevlyar/graphql-go/introspection"
	"io"
	"sort"
	"strings"
)

const privateTypeMatcher string = `__`

var privateTypeMatcherLen int = len(privateTypeMatcher)

var declareOfArgsType = make(map[string]string, 0)
var declareOfArgsNames = make([]string, 0)

func addForDeclareNewArgsType(name string, definition string) {
	exists, ok := declareOfArgsType[name]
	if ok {
		if exists != definition {
			panic(
				fmt.Sprintf("try to set args type with one name and with other definitions \n\tname=%s,\n\told='%s',\n\tnew='%s'",
					name,
					exists,
					definition,
				),
			)
		}
	} else {
		declareOfArgsNames = append(declareOfArgsNames, name)
		declareOfArgsType[name] = definition
	}
}

func getNameOfType(t *introspection.Type, withOutSufix bool, innerCall bool, forOutOrDeclare bool, hasDefaultValue bool, callFromMain bool) (nameStr string, needDeclare bool, baseType bool) {
	var (
		sufix  string
		prefix string
	)

	if !withOutSufix {
		sufix = `Resolver`
	}

	if t == nil {
		return ``, false, false
	}

	name := t.Name()

	switch t.Kind() {
	case `ENUM`:
		//nameOfEnum := t.Name()
		//if callFromMain && nameOfEnum != nil {
		//	return *nameOfEnum, true, false
		//}
		if !innerCall && !callFromMain && !hasDefaultValue {
			prefix = `*`
		}
		sufix = ``

		//return prefix + *nameOfEnum, false, true
	case `INTERFACE`, `UNION`:
		if forOutOrDeclare {
			prefix = `*`
		}
	case `INPUT_OBJECT`:
		if !innerCall && !forOutOrDeclare {
			prefix = `*`
		}
	case `SCALAR`:
		if !innerCall && !hasDefaultValue {
			prefix = `*`
		}
	case `LIST`:
		if innerCall {
			prefix = `[]`
		} else {
			prefix = `*[]`
		}
	}

	if name == nil {
		if t.Kind() == `LIST` {
			innerCall = false
		} else {
			innerCall = true
		}

		nameStr, needDeclare, baseType = getNameOfType(t.OfType(), withOutSufix, innerCall, forOutOrDeclare, hasDefaultValue, false)
		nameStr = prefix + nameStr

		return nameStr, needDeclare, baseType
	}

	switch *name {
	case `ID`:
		baseType = true
		nameStr = prefix + `graphql.ID`
	case `Time`:
		baseType = true
		nameStr = prefix + `graphql.Time`
	case `String`:
		baseType = true
		nameStr = prefix + `string`
	case `Float`:
		baseType = true
		nameStr = prefix + `float64`
	case `Boolean`:
		baseType = true
		nameStr = prefix + `bool`
	case `Int`:
		baseType = true
		nameStr = prefix + `int32`
	default:
		if !(len(*name) > 2 && (*name)[0:privateTypeMatcherLen] == privateTypeMatcher) {
			nameStr = prefix + *name + sufix
			needDeclare = true
			baseType = false
		}
	}
	return nameStr, needDeclare, baseType
}

func fmtName(name string) string {
	if strings.ToUpper(name) == `ID` {
		return `ID`
	}
	return strings.Title(name)
}

func printArgs(funcName string, f *introspection.Field) (declareArg string) {
	if len(f.Args()) == 0 {
		return declareArg
	}

	buf := &bytes.Buffer{}

	buf.WriteString(fmt.Sprintf("type %sArguments struct {\n", funcName))

	for _, input := range f.Args() {
		nameArg := input.Name()
		typeOf := input.Type()

		printDescription("\t", buf, input)

		var hasDefValue bool
		if v := input.DefaultValue(); v != nil {
			hasDefValue = true
			buf.WriteString(fmt.Sprintf("\t// default value \"%s\"\n", *v))
		}
		// just input - is struct - not need %sResolver
		nameOfType, _, _ := getNameOfType(typeOf, true, false, false, hasDefValue, false)
		buf.WriteString(fmt.Sprintf("\t%s %s\n", fmtName(nameArg), nameOfType))
	}

	buf.WriteString("}\n\n")

	declareArg = fmt.Sprintf("in %sArguments", funcName)
	addForDeclareNewArgsType(funcName+"Arguments", buf.String())

	return declareArg

}

func printInputValue(prefix string, buf *bytes.Buffer, f *introspection.InputValue) {
	printDescription(prefix, buf, f)

	name := fmtName(f.Name())

	var hasDefValue bool
	if dv := f.DefaultValue(); dv != nil {
		hasDefValue = true
		_, _ = buf.WriteString(fmt.Sprintf("%s// default value - \"%s\"\n", prefix, *dv))
	}
	typeStr, _, _ := getNameOfType(f.Type(), true, false, false, hasDefValue, false)

	_, _ = buf.WriteString(prefix + name + ` ` + typeStr + "\n")
}

func printField(tBuf *bytes.Buffer, f *introspection.Field) {
	printDescription("\t", tBuf, f)
	printDeprecatedInfo("\t", tBuf, f)

	// write name of function
	funcName := fmtName(f.Name())

	// write args
	args := printArgs(funcName, f)
	// find out
	var data string
	outType := f.Type()
	var (
		nameOut  string
		baseType bool
	)
	if outType != nil {
		nameOut, _, baseType = getNameOfType(outType, false, false, true, false, false)

		desc := f.Description()
		if !baseType || (desc != nil && strings.Contains(*desc, "@lazy"))  {
			if args == "" {
				args = "ctx context.Context"
			} else {
				args = "ctx context.Context, " + args
			}

			data = "(" + nameOut + ", error)"
		} else {
			data = nameOut
		}
	}

	tBuf.WriteString(fmt.Sprintf("\t%s(%s) %s\n", funcName, args, data))
}

func Print(sourceName, packageName, schema string, writer io.Writer) error {

	obj, err := graphql.ParseSchema(schema, nil)
	if err != nil {
		return err
	}

	headerBuf := &bytes.Buffer{}

	IncludeDeprecated := &struct {
		IncludeDeprecated bool
	}{
		IncludeDeprecated: true,
	}
	descriptor := obj.Inspect()
	declareBuf := &bytes.Buffer{}
	tBuf := &bytes.Buffer{}
	typeDesc := descriptor.Types()
	for _, td := range typeDesc {

		nameOfType, needDeclare, _ := getNameOfType(td, false, false, false, false, true)
		if !needDeclare {
			continue
		}
		tBuf.WriteString("\n")
		tBuf.WriteString("\n")

		printDescription(``, tBuf, td)

		switch td.Kind() {
		case "ENUM":
			// set alias for string
			values := td.EnumValues(IncludeDeprecated)
			if values != nil {
				tBuf.WriteString(fmt.Sprintf("type %s = string\n", nameOfType))
				tBuf.WriteString("const (\n")
				for _, enumValue := range *values {
					printDescription("\t", tBuf, enumValue)
					printDeprecatedInfo("\t", tBuf, enumValue)
					enumStr := strings.ToUpper(enumValue.Name())
					tBuf.WriteString(fmt.Sprintf("\t%s_%s = `%s`\n", nameOfType, enumStr, enumStr))
				}
				tBuf.WriteString(")\n")
			}
		case "UNION":
			// generate implementation for switch
			mayBeRealisation := td.PossibleTypes()
			if mayBeRealisation != nil {
				tBuf.WriteString(fmt.Sprintf("type %s struct {\n", nameOfType))
				tBuf.WriteString(fmt.Sprintf("\tResult interface{}\n"))
				tBuf.WriteString("}\n\n")

				for _, realT := range *mayBeRealisation {
					nameImplement, _, _ := getNameOfType(realT, true, false, false, false, false)
					nameImplementResolver, _, _ := getNameOfType(realT, false, false, false, false, false)
					tBuf.WriteString(fmt.Sprintf("func (i *%s) To%s() (%s, bool) {\n", nameOfType, nameImplement, nameImplementResolver))
					tBuf.WriteString(fmt.Sprintf("\tc, ok := i.Result.(%s)\n\treturn c, ok\n", nameImplementResolver))
					tBuf.WriteString("}\n\n")
				}
			}

		case "INTERFACE":
			nameOfInterface, _, _ := getNameOfType(td, true, false, false, false, false)

			tBuf.WriteString(fmt.Sprintf("type %s interface {\n", nameOfInterface))
			fields := td.Fields(IncludeDeprecated)
			if fields != nil {
				for _, f := range *fields {
					printField(tBuf, f)
				}
			}
			tBuf.WriteString("}\n\n")

			// generate implementation for switch
			mayBeRealisation := td.PossibleTypes()
			if mayBeRealisation != nil {
				tBuf.WriteString(fmt.Sprintf("type %s struct {\n", nameOfType))
				tBuf.WriteString(fmt.Sprintf("\t%s\n", nameOfInterface))
				tBuf.WriteString("}\n\n")

				for _, realT := range *mayBeRealisation {
					nameImplement, _, _ := getNameOfType(realT, true, false, false, false, false)
					nameImplementResolver, _, _ := getNameOfType(realT, false, false, false, false, false)
					tBuf.WriteString(fmt.Sprintf("func (i *%s) To%s() (%s, bool) {\n", nameOfType, nameImplement, nameImplementResolver))
					tBuf.WriteString(fmt.Sprintf("\tc, ok := i.%s.(%s)\n\treturn c, ok\n", nameOfInterface, nameImplementResolver))
					tBuf.WriteString("}\n\n")
				}
			}

		case "OBJECT":
			tBuf.WriteString(fmt.Sprintf("type %s interface {\n", nameOfType))
			fields := td.Fields(IncludeDeprecated)
			if fields != nil {
				for _, f := range *fields {
					printField(tBuf, f)
				}
			}
			tBuf.WriteString("}\n\n")
		case "INPUT_OBJECT":
			nameOfInput, _, _ := getNameOfType(td, true, false, true, false, false)
			tBuf.WriteString(fmt.Sprintf("type %s struct {\n", nameOfInput))
			inputValues := td.InputFields()
			if inputValues != nil {
				for _, f := range *inputValues {
					printInputValue("\t", tBuf, f)
				}
			}
			tBuf.WriteString("}\n\n")

		}
	}

	headerBuf.WriteString(fmt.Sprintf("// Code generated by gql-gen-resolver. DO NOT EDIT.\n// source: %s\npackage %s\n", sourceName, packageName))

	headerBuf.WriteString(`
import (
	graphql "github.com/sevlyar/graphql-go"
	context "context"
)
`)

	headerBuf.WriteString(fmt.Sprintf("\n// schema from source: %s \nconst Schema string = `\n%s\n`", sourceName, schema))

	headerBuf.WriteString("\n\ntype SchemaResolver interface {\n")
	if descriptor.QueryType() != nil {
		headerBuf.WriteString("\tQueryResolver\n")
	}

	if descriptor.MutationType() != nil {
		headerBuf.WriteString("\tMutationResolver\n")
	}

	if descriptor.SubscriptionType() != nil {
		headerBuf.WriteString("\tSubscriptionResolver\n")
	}
	headerBuf.WriteString("}\n\n")

	headerBuf.WriteTo(writer)

	sort.Strings(declareOfArgsNames)

	for _, key := range declareOfArgsNames {
		declareBuf.WriteString(declareOfArgsType[key])
	}

	declareBuf.WriteTo(writer)

	tBuf.WriteTo(writer)

	if err != nil {
		return err
	}

	return nil
}

type Deprecator interface {
	IsDeprecated() bool
	DeprecationReason() *string
}

func printDeprecatedInfo(prefix string, buf *bytes.Buffer, field Deprecator) {
	if field == nil || buf == nil {
		return
	}

	if field.IsDeprecated() {
		depStr := prefix + "// Deprecated:"
		/*
			//
			// Deprecated: Not implemented, do not use.
		*/
		reason := field.DeprecationReason()
		if reason != nil {
			depStr += " " + *reason
		}
		buf.WriteString(depStr + "\n")
	}
}

type Descriptor interface {
	Description() *string
}

func printDescription(prefix string, buf *bytes.Buffer, t Descriptor) {
	if t == nil || buf == nil {
		return
	}

	desc := t.Description()
	prefix += "// "
	if desc != nil {
		for _, msg := range strings.Split(*desc, "\n") {
			buf.WriteString(prefix + msg + "\n")
		}
	}
}
