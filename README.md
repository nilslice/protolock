### Work In Progress

# protolock

Track your `.proto` files and prevent incompatible changes to field names and numbers.

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

---

## Acknowledgement

Thank you to Ernest Micklei for his work on the excellent parser heavily relied upon by this tool and many more: [https://github.com/emicklei/proto](https://github.com/emicklei/proto)