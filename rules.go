package protolock

import "fmt"

var (
	// ruleFuncs provides a complete list of all funcs to be run to compare
	// a set of Protolocks. This list should be updated as new RuleFunc's
	// are added to this package.
	ruleFuncs = []RuleFunc{
		NoUsingReservedFields,
		NoRemovingReservedFields,
		NoChangingFieldIDs,
		NoChangingFieldTypes,
		NoChangingFieldNames,
		NoRemovingRPCs,
	}

	strict = true
	debug  = false
)

// SetStrict enables the user to toggle strict mode on and off.
func SetStrict(mode bool) {
	strict = mode
}

// SetDebug enables the user to toggle debug mode on and off.
func SetDebug(status bool) {
	debug = status
}

// RuleFunc defines the common signature for a function which can compare
// Protolock states and determine if issues exist.
type RuleFunc func(current, updated Protolock) ([]Warning, bool)

// lockIDsMap:
// table of filepath -> message name -> reserved field ID -> times ID encountered
// i.e.
/*
	["test.proto"] 	-> ["Test"] -> [1] -> 1

			-> ["User"] -> [1] -> 1
				       [2] -> 1
				       [3] -> 1

			-> ["Plan"] -> [1] -> 1
				       [2] -> 1
				       [3] -> 1
*/
type lockIDsMap map[string]map[string]map[int]int

// lockNamesMap:
// table of filepath -> message name -> field name -> times name encountered (or the field ID)
// i.e.
/*
	["test.proto"]	->	["Test"]	->	["field_one"]	->	1
			-> 	["User"] 	-> 	["field_one"] 	-> 	1
							["field_two"] 	-> 	1
							["field_three"] -> 	1

			-> 	["Plan"] 	-> 	["field_one"] 	-> 	1
				       			["field_two"] 	-> 	1
				       			["field_three"] -> 	1
			# if mapping field name -> id,
			-> 	["Account"] 	-> 	["field_one"] 	-> 	1
						-> 	["field_two"] 	-> 	2
						-> 	["field_three"] -> 	3
*/
type lockNamesMap map[string]map[string]map[string]int

// lockFieldMap:
// table of filepath -> message name -> field name -> field type
type lockFieldMap map[string]map[string]map[string]Field

// lockFieldIDNameMap:
// table of filepath -> message name -> field ID -> field name
type lockFieldIDNameMap map[string]map[string]map[int]string

// NoUsingReservedFields compares the current vs. updated Protolock definitions
// and will return a list of warnings if any message's previously reserved fields
// are now being used as part of the same message.
func NoUsingReservedFields(cur, upd Protolock) ([]Warning, bool) {
	if debug {
		beginRuleDebug("NoUsingReservedFields")
	}

	reservedIDMap, reservedNameMap := getReservedFields(cur)

	// add each messages field name/number to the existing list identified as
	// reserved to analyze
	for _, def := range upd.Definitions {
		if reservedIDMap[def.Filepath] == nil {
			reservedIDMap[def.Filepath] = make(map[string]map[int]int)
		}
		if reservedNameMap[def.Filepath] == nil {
			reservedNameMap[def.Filepath] = make(map[string]map[string]int)
		}
		for _, msg := range def.Def.Messages {
			for _, field := range msg.Fields {
				if reservedIDMap[def.Filepath][msg.Name] == nil {
					reservedIDMap[def.Filepath][msg.Name] = make(map[int]int)
				}
				if reservedNameMap[def.Filepath][msg.Name] == nil {
					reservedNameMap[def.Filepath][msg.Name] = make(map[string]int)
				}
				reservedIDMap[def.Filepath][msg.Name][field.ID]++
				reservedNameMap[def.Filepath][msg.Name][field.Name]++
			}
		}
	}

	var warnings []Warning
	// if the field ID was encountered more than once per message, then it
	// is known to be a re-use of a reserved field and a warning should be
	// returned for each occurrance
	for path, m := range reservedIDMap {
		for msgName, mm := range m {
			for id, count := range mm {
				if count > 1 {
					msg := fmt.Sprintf(
						"%s is re-using ID: %d, a reserved field",
						msgName, id,
					)
					warnings = append(warnings, Warning{
						Filepath: path,
						Message:  msg,
					})
				}
			}
		}
	}
	// if the field name was encountered more than once per message, then it
	// is known to be a re-use of a reserved field and a warning should be
	// returned for each occurrance
	for path, m := range reservedNameMap {
		for msgName, mm := range m {
			for name, count := range mm {
				if count > 1 {
					msg := fmt.Sprintf(
						`%s is re-using name: "%s", a reserved field`,
						msgName, name,
					)
					warnings = append(warnings, Warning{
						Filepath: path,
						Message:  msg,
					})
				}
			}
		}
	}

	if debug {
		concludeRuleDebug("NoUsingReservedFields", warnings)
	}

	if warnings != nil {
		return warnings, false
	}

	return nil, true
}

// NoRemovingReservedFields compares the current vs. updated Protolock definitions
// and will return a list of warnings if any reserved field has been removed. This
// rule is only enforced when strict mode is enabled.
func NoRemovingReservedFields(cur, upd Protolock) ([]Warning, bool) {
	if !strict {
		return nil, true
	}

	if debug {
		beginRuleDebug("NoRemovingReservedFields")
	}

	var warnings []Warning
	// check that all reserved fields on current Protolock remain in the
	// updated Protolock
	curReservedIDMap, curReservedNameMap := getReservedFields(cur)
	updReservedIDMap, updReservedNameMap := getReservedFields(upd)
	for path, msgMap := range curReservedIDMap {
		for msgName, idMap := range msgMap {
			for id := range idMap {
				if _, ok := updReservedIDMap[path][msgName][id]; !ok {
					msg := fmt.Sprintf(
						"%s is missing ID: %d, a reserved field",
						msgName, id,
					)
					warnings = append(warnings, Warning{
						Filepath: path,
						Message:  msg,
					})
				}
			}
		}
	}
	for path, msgMap := range curReservedNameMap {
		for msgName, nameMap := range msgMap {
			for name := range nameMap {
				if _, ok := updReservedNameMap[path][msgName][name]; !ok {
					msg := fmt.Sprintf(
						`%s is missing name: "%s", a reserved field`,
						msgName, name,
					)
					warnings = append(warnings, Warning{
						Filepath: path,
						Message:  msg,
					})
				}
			}
		}
	}

	if debug {
		concludeRuleDebug("NoRemovingReservedFields", warnings)
	}

	if warnings != nil {
		return warnings, false
	}

	return nil, true
}

// NoChangingFieldIDs compares the current vs. updated Protolock definitions and
// will return a list of warnings if any field ID number has been changed.
func NoChangingFieldIDs(cur, upd Protolock) ([]Warning, bool) {
	if debug {
		beginRuleDebug("NoChangingFieldIDs")
	}

	curNameIDMap := getNonReservedFields(cur)
	updNameIDMap := getNonReservedFields(upd)

	var warnings []Warning
	// check that all current Protolock names map to the same IDs as the
	// updated Protolock
	for path, msgMap := range curNameIDMap {
		for msgName, fieldMap := range msgMap {
			for fieldName, fieldID := range fieldMap {
				updFieldID, ok := updNameIDMap[path][msgName][fieldName]
				if ok {
					if updFieldID != fieldID {
						msg := fmt.Sprintf(
							`%s field: "%s" has a different ID: %d, previously %d`,
							msgName, fieldName, updFieldID, fieldID,
						)
						warnings = append(warnings, Warning{
							Filepath: path,
							Message:  msg,
						})
					}
				}
			}
		}
	}

	if debug {
		concludeRuleDebug("NoChangingFieldIDs", warnings)
	}

	if warnings != nil {
		return warnings, false
	}

	return nil, true
}

// NoChangingFieldTypes compares the current vs. updated Protolock definitions and
// will return a list of warnings if any field type has been changed.
func NoChangingFieldTypes(cur, upd Protolock) ([]Warning, bool) {
	if debug {
		beginRuleDebug("NoChangingFieldTypes")
	}

	curFieldMap := getFieldMap(cur)
	updFieldMap := getFieldMap(upd)
	var warnings []Warning
	// check that the current Protolock message's field types are the same
	// for each of the same message's fields in the updated Protolock
	for path, msgMap := range curFieldMap {
		for msgName, fieldMap := range msgMap {
			for fieldName, field := range fieldMap {
				updField, ok := updFieldMap[path][msgName][fieldName]
				if ok {
					if updField.Type != field.Type {
						msg := fmt.Sprintf(
							`%s field: "%s" has a different type: %s, previously %s`,
							msgName, fieldName, updField.Type, field.Type,
						)
						warnings = append(warnings, Warning{
							Filepath: path,
							Message:  msg,
						})
					}

					if updField.IsRepeated != field.IsRepeated {
						msg := fmt.Sprintf(
							`%s field: "%s" has a different "repeated" status: %t, previously %t`,
							msgName, fieldName, updField.IsRepeated, field.IsRepeated,
						)
						warnings = append(warnings, Warning{
							Filepath: path,
							Message:  msg,
						})
					}
				}
			}
		}
	}

	if debug {
		concludeRuleDebug("NoChangingFieldTypes", warnings)
	}

	if warnings != nil {
		return warnings, false
	}

	return nil, true
}

// NoChangingFieldNames compares the current vs. updated Protolock definitions and
// will return a list of warnings if any message's previous fields have been
// renamed. This rule is only enforced when strict mode is enabled.
func NoChangingFieldNames(cur, upd Protolock) ([]Warning, bool) {
	if debug {
		beginRuleDebug("NoChangingFieldNames")
	}

	if !strict {
		return nil, true
	}

	curFieldMap := getFieldsIDName(cur)
	updFieldMap := getFieldsIDName(upd)

	var warnings []Warning
	// check that the current Protolock messages' field names are equal to
	// their relative messages' field names in the updated Protolock
	for path, msgMap := range curFieldMap {
		for msgName, fieldMap := range msgMap {
			for fieldID, fieldName := range fieldMap {
				updFieldName, ok := updFieldMap[path][msgName][fieldID]
				if ok {
					if updFieldName != fieldName {
						msg := fmt.Sprintf(
							`%s field: "%s" (ID: %d) has an updated name, previously "%s"`,
							msgName, updFieldName, fieldID, fieldName,
						)
						warnings = append(warnings, Warning{
							Filepath: path,
							Message:  msg,
						})
					}
				}
			}
		}
	}

	if debug {
		concludeRuleDebug("NoChangingFieldNames", warnings)
	}

	if warnings != nil {
		return warnings, false
	}

	return nil, true
}

// NoRemovingRPCs compares the current vs. updated Protolock definitions and
// will return a list of warnings if any RPCs provided by a Service have been
// removed. This rule is only enforced when strict mode is enabled.
func NoRemovingRPCs(cur, upd Protolock) ([]Warning, bool) {
	if !strict {
		return nil, true
	}
	return nil, true
}

// NoDeprecatingFieldsWithoutReserve compares the current vs. updated Protolock
// definitions and will return a list of warnings if any field has been removed
// without a corresponding reservation of that field or name.
func NoDeprecatingFieldsWithoutReserve(cur, upd Protolock) ([]Warning, bool) {
	if !strict {

	}
	return nil, true
}

// getReservedFields gets all the reserved field numbers and names, and stashes
// them in a lockIDsMap and lockNamesMap to be checked against.
func getReservedFields(lock Protolock) (lockIDsMap, lockNamesMap) {
	reservedIDMap := make(lockIDsMap)
	reservedNameMap := make(lockNamesMap)

	for _, def := range lock.Definitions {
		if reservedIDMap[def.Filepath] == nil {
			reservedIDMap[def.Filepath] = make(map[string]map[int]int)
		}
		if reservedNameMap[def.Filepath] == nil {
			reservedNameMap[def.Filepath] = make(map[string]map[string]int)
		}
		for _, msg := range def.Def.Messages {
			for _, id := range msg.ReservedIDs {
				if reservedIDMap[def.Filepath][msg.Name] == nil {
					reservedIDMap[def.Filepath][msg.Name] = make(map[int]int)
				}
				reservedIDMap[def.Filepath][msg.Name][id]++
			}
			for _, name := range msg.ReservedNames {
				if reservedNameMap[def.Filepath][msg.Name] == nil {
					reservedNameMap[def.Filepath][msg.Name] = make(map[string]int)
				}
				reservedNameMap[def.Filepath][msg.Name][name]++
			}
		}
	}

	return reservedIDMap, reservedNameMap
}

func getFieldsIDName(lock Protolock) lockFieldIDNameMap {
	fieldIDNameMap := make(lockFieldIDNameMap)

	for _, def := range lock.Definitions {
		if fieldIDNameMap[def.Filepath] == nil {
			fieldIDNameMap[def.Filepath] = make(map[string]map[int]string)
		}
		for _, msg := range def.Def.Messages {
			for _, field := range msg.Fields {
				if fieldIDNameMap[def.Filepath][msg.Name] == nil {
					fieldIDNameMap[def.Filepath][msg.Name] = make(map[int]string)
				}
				fieldIDNameMap[def.Filepath][msg.Name][field.ID] = field.Name
			}
		}
	}

	return fieldIDNameMap
}

// getNonReservedFields gets all the reserved field numbers and names, and stashes
// them in a lockNamesMap to be checked against.
func getNonReservedFields(lock Protolock) lockNamesMap {
	nameIDMap := make(lockNamesMap)

	for _, def := range lock.Definitions {
		if nameIDMap[def.Filepath] == nil {
			nameIDMap[def.Filepath] = make(map[string]map[string]int)
		}
		for _, msg := range def.Def.Messages {
			for _, field := range msg.Fields {
				if nameIDMap[def.Filepath][msg.Name] == nil {
					nameIDMap[def.Filepath][msg.Name] = make(map[string]int)
				}
				nameIDMap[def.Filepath][msg.Name][field.Name] = field.ID
			}
		}
	}

	return nameIDMap
}

// getFieldMap gets all the field names and types, and stashes them in a
// lockTypesMap to be checked against.
func getFieldMap(lock Protolock) lockFieldMap {
	nameTypeMap := make(lockFieldMap)

	for _, def := range lock.Definitions {
		if nameTypeMap[def.Filepath] == nil {
			nameTypeMap[def.Filepath] = make(map[string]map[string]Field)
		}
		for _, msg := range def.Def.Messages {
			for _, field := range msg.Fields {
				if nameTypeMap[def.Filepath][msg.Name] == nil {
					nameTypeMap[def.Filepath][msg.Name] = make(map[string]Field)
				}
				nameTypeMap[def.Filepath][msg.Name][field.Name] = field
			}
		}
	}

	return nameTypeMap
}

func beginRuleDebug(name string) {
	fmt.Println("RUN RULE:", name)
}

func concludeRuleDebug(name string, warnings []Warning) {
	fmt.Println("# Warnings:", len(warnings))
	for i, w := range warnings {
		msg := fmt.Sprintf("%d). %s [%s]", i+1, w.Message, w.Filepath)
		fmt.Println(msg)
	}
	fmt.Println("END RULE:", name)
	fmt.Println("===")
}
