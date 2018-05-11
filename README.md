# protolock

Track your `.proto` files and prevent incompatible changes to field names and numbers.

[![CircleCI](https://circleci.com/gh/nilslice/protolock/tree/master.svg?style=svg)](https://circleci.com/gh/nilslice/protolock/tree/master)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/nilslice/protolock)

## Why

Ever _accidentally_ break your API compatibility while you're busy fixing problems? You may have forgotten to reserve the field number of a message or you re-ordered fields after removing a property. Maybe a new team member was not familiar with the backward-compatibility of Protocol Buffers and made an easy mistake.

`protolock` attempts to help prevent this from happening.

## Usage

Similar in concept to the higher-level features of `git`, track and/or prevent changes to your `.proto` files. 

1. **Initialize** your repository: 

        $ protolock init
        # creates a `proto.lock` file

3. **Add changes** to .proto messages or services, verify no breaking changes made: 

        $ protolock status

2. **Commit** a new state of your .protos (overwrites `proto.lock`): 

        $ protolock commit

4. **Integrate** into your protobuf compilation step: 

        $ protolock status && protoc --I ...

In all, prevent yourself from compiling your protobufs and generating code if breaking changes have been made.

**Recommended:** commit the output `proto.lock` file into your version control system

## Rules

### No Using Reserved Fields
Compares the current vs. updated Protolock definitions and will return a list of 
warnings if any message's previously reserved fields or IDs are now being used 
as part of the same message.

### No Removing Reserved Fields
Compares the current vs. updated Protolock definitions and will return a list of 
warnings if any reserved field has been removed. 

**Note:** This rule is only enforced when strict mode is enabled.


### No Changing Field IDs
Compares the current vs. updated Protolock definitions and will return a list of 
warnings if any field ID number has been changed.


### No Changing Field Types
Compares the current vs. updated Protolock definitions and will return a list of 
warnings if any field type has been changed.


### No Changing Field Names
Compares the current vs. updated Protolock definitions and will return a list of 
warnings if any message's previous fields have been renamed. 

**Note:** This rule is only enforced when strict mode is enabled. 

### No Removing Fields Without Reserve
Compares the current vs. updated Protolock definitions and will return a list of 
warnings if any field has been removed without a corresponding reservation of 
that field or name.

### No Removing RPCs
Compares the current vs. updated Protolock definitions and will return a list of 
warnings if any RPCs provided by a Service have been removed. 

**Note:** This rule is only enforced when strict mode is enabled. 

### No Changing RPC Signature
Compares the current vs. updated Protolock definitions and will return a list of 
warnings if any RPC signature has been changed while using the same name.

---

## Acknowledgement

Thank you to Ernest Micklei for his work on the excellent parser heavily relied upon by this tool and many more: [https://github.com/emicklei/proto](https://github.com/emicklei/proto)