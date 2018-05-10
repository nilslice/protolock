### Work In Progress

# protolock

Track your `.proto` files and prevent incompatible changes to field names and numbers.

## Why

Ever _accidentally_ break your API compatibility while you're busy fixing problems? You may have forgotten to reserve the field number of a message or re-order fields after removing a property. A new team member may not be familiar with the backward-compatibility of Protocol Buffers and make an easy mistake. 

`protolock` attempts to help prevent this from happening.

## Usage

Similar in concept to the higher-level features of `git`, track and/or prevent changes to your `.proto` files. 

1. Initialize your repository: 

        $ protolock init

3. Add changes to protobuf messages or services: 

        $ protolock commit

2. Check that no breaking changes were made: 

        $ protolock status

4. Integrate into your protobuf compilation step: 

        $ protolock status && protoc --I ...

In all, prevent yourself from compiling your protobufs and generating code if breaking changes have been made.

---
**Recommended:** commit the output `proto.lock` file into your version control system

---
