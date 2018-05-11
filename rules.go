package protolock

import "fmt"

var (
	// ruleFuncs provides a complete list of all funcs to be run to compare
	// a set of Protolocks. This list should be updated as new RuleFunc's
	// are added to this package.
	ruleFuncs = []RuleFunc{
		NoUsingReservedFields,
		NoRemovingReservedFields,
		NoRemovingFieldsWithoutReserve,
		NoChangingFieldIDs,
		NoChangingFieldTypes,
		NoChangingFieldNames,
		NoRemovingRPCs,
		NoChangingRPCSignature,
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

// lockRPCMap:
// table of filepath -> service name -> rpc name -> rpc type
type lockRPCMap map[string]map[string]map[string]RPC

// lockFieldIDNameMap:
// table of filepath -> message name -> field ID -> field name
type lockFieldIDNameMap map[string]map[string]map[int]string

// NoUsingReservedFields compares the current vs. updated Protolock definitions
// and will return a list of warnings if any message's previously reserved fields
// or IDs are now being used as part of the same message.
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
						`"%s" is re-using ID: %d, a reserved field`,
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
						`"%s" is re-using name: "%s", a reserved field`,
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
						`"%s" is missing ID: %d, which had been reserved`,
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
						`"%s" is missing name: "%s", which had been reserved`,
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
							`"%s" field: "%s" has a different ID: %d, previously %d`,
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
							`"%s" field: "%s" has a different type: %s, previously %s`,
							msgName, fieldName, updField.Type, field.Type,
						)
						warnings = append(warnings, Warning{
							Filepath: path,
							Message:  msg,
						})
					}

					if updField.IsRepeated != field.IsRepeated {
						msg := fmt.Sprintf(
							`"%s" field: "%s" has a different "repeated" status: %t, previously %t`,
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
	if !strict {
		return nil, true
	}

	if debug {
		beginRuleDebug("NoChangingFieldNames")
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
							`"%s" field: "%s" ID: %d has an updated name, previously "%s"`,
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

	if debug {
		beginRuleDebug("NoRemovingRPCs")
	}

	var warnings []Warning
	// check that all current Protolock services' RPCs are still in the
	// updated Protolock
	curServices := getServicesRPCsMap(cur)
	updServices := getServicesRPCsMap(upd)

	for path, svcMap := range curServices {
		for svcName, rpcMap := range svcMap {
			for rpcName := range rpcMap {
				_, ok := updServices[path][svcName][rpcName]
				if !ok {
					msg := fmt.Sprintf(
						`"%s" is missing RPC: "%s", which should be available`,
						svcName, rpcName,
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
		concludeRuleDebug("NoRemovingRPCs", warnings)
	}

	if warnings != nil {
		return warnings, false
	}

	return nil, true
}

// NoRemovingFieldsWithoutReserve compares the current vs. updated Protolock
// definitions and will return a list of warnings if any field has been removed
// without a corresponding reservation of that field or name.
func NoRemovingFieldsWithoutReserve(cur, upd Protolock) ([]Warning, bool) {
	if debug {
		beginRuleDebug("NoRemovingFieldsWithoutReserve")
	}

	var warnings []Warning
	// check that if a field name from the current Protolock is not retained
	// in the updated Protolock, then the field's name and ID shoule become
	// reserved within the parent message
	curFieldMap := getFieldMap(cur)
	updFieldMap := getFieldMap(upd)
	for path, msgMap := range curFieldMap {
		for msgName, fieldMap := range msgMap {
			for fieldName, field := range fieldMap {
				_, ok := updFieldMap[path][msgName][fieldName]
				if !ok {
					// check that the field name and ID are
					// both in the reserved fields for this
					// message
					resIDsMap, resNamesMap := getReservedFields(upd)
					if _, ok := resIDsMap[path][msgName][field.ID]; !ok {
						msg := fmt.Sprintf(
							`"%s" ID: "%d" has been removed, but is not "reserved"`,
							msgName, field.ID,
						)
						warnings = append(warnings, Warning{
							Filepath: path,
							Message:  msg,
						})
					}
					if _, ok := resNamesMap[path][msgName][field.Name]; !ok {
						msg := fmt.Sprintf(
							`"%s" field: "%s" has been removed, but is not "reserved"`,
							msgName, field.Name,
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
		concludeRuleDebug("NoRemovingFieldsWithoutReserve", warnings)
	}

	if warnings != nil {
		return warnings, false
	}

	return nil, true
}

// NoChangingRPCSignature compares the current vs. updated Protolock
// definitions and will return a list of warnings if any RPC signature has been
// changed while using the same name.
func NoChangingRPCSignature(cur, upd Protolock) ([]Warning, bool) {
	if debug {
		beginRuleDebug("NoChangingRPCSignature")
	}

	var warnings []Warning
	// check that no breaking changes to the signature of an RPC have been
	// made between the current Protolock and the updated Protolock
	curRPCMap := getRPCMap(cur)
	updRPCMap := getRPCMap(upd)
	for path, svcMap := range curRPCMap {
		for svcName, rpcMap := range svcMap {
			for rpcName, rpc := range rpcMap {
				updRPC, ok := updRPCMap[path][svcName][rpcName]
				if !ok {
					continue
				}

				// check that stream option and type are the same
				// for both the RPC's request and response
				if rpc.InStreamed != updRPC.InStreamed {
					msg := fmt.Sprintf(
						`"%s" RPC: "%s" input stream identifier has changed, previously: %t`,
						svcName, rpcName, rpc.InStreamed,
					)
					warnings = append(warnings, Warning{
						Filepath: path,
						Message:  msg,
					})
				}

				if rpc.OutStreamed != updRPC.OutStreamed {
					msg := fmt.Sprintf(
						`"%s" RPC: "%s" output stream identifier has changed, previously: %t`,
						svcName, rpcName, rpc.OutStreamed,
					)
					warnings = append(warnings, Warning{
						Filepath: path,
						Message:  msg,
					})
				}

				if rpc.InType != updRPC.InType {
					msg := fmt.Sprintf(
						`"%s" RPC: "%s" input type has changed, previously: %s`,
						svcName, rpcName, rpc.InType,
					)
					warnings = append(warnings, Warning{
						Filepath: path,
						Message:  msg,
					})
				}

				if rpc.OutType != updRPC.OutType {
					msg := fmt.Sprintf(
						`"%s" RPC: "%s" output type has changed, previously: %s`,
						svcName, rpcName, rpc.OutType,
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
		concludeRuleDebug("NoChangingRPCSignature", warnings)
	}

	if warnings != nil {
		return warnings, false
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

// getFieldsIDName gets all the fields mapped by the field ID to its name for
// all messages.
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
// lockFieldMap to be checked against.
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

// getServicesRPCsMap gets all the RPCs for the Services in a Protolock and
// stashes them in a lockNamesMap to be checked against.
func getServicesRPCsMap(lock Protolock) lockNamesMap {
	servicesRPCsMap := make(lockNamesMap)
	for _, def := range lock.Definitions {
		if servicesRPCsMap[def.Filepath] == nil {
			servicesRPCsMap[def.Filepath] = make(map[string]map[string]int)
		}
		for _, svc := range def.Def.Services {
			if servicesRPCsMap[def.Filepath][svc.Name] == nil {
				servicesRPCsMap[def.Filepath][svc.Name] = make(map[string]int)
			}
			for _, rpc := range svc.RPCs {
				servicesRPCsMap[def.Filepath][svc.Name][rpc.Name]++
			}
		}
	}

	return servicesRPCsMap
}

// getRPCMap gets all the RPC names and types, and stashes them in a
// lockRPCMap to be checked against.
func getRPCMap(lock Protolock) lockRPCMap {
	rpcTypeMap := make(lockRPCMap)

	for _, def := range lock.Definitions {
		if rpcTypeMap[def.Filepath] == nil {
			rpcTypeMap[def.Filepath] = make(map[string]map[string]RPC)
		}
		for _, svc := range def.Def.Services {
			for _, rpc := range svc.RPCs {
				if rpcTypeMap[def.Filepath][svc.Name] == nil {
					rpcTypeMap[def.Filepath][svc.Name] = make(map[string]RPC)
				}
				rpcTypeMap[def.Filepath][svc.Name][rpc.Name] = rpc
			}
		}
	}

	return rpcTypeMap
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
