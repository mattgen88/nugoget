# NuGoGet

Nuget Go Get, because it's dumb I can't just easily update all my deps from the
command line.

## Warning
I really didn't put much work into this. Under the hood it just runs dotnet
commands and parses the output. Will likely break with dotnet updates. Hopefully
won't break when updating dependencies because it runs dotnet commands to
install the appropriate version.

## Dependencies
- go 1.15.8+, probably a lot lower too
- dotnet 5, probably not lower, probably not higher.

## Built With
- [github.com/spf13/cobra](https://github.com/spf13/cobra)

## Installation
`go get github.com/mattgen88/nugoget`


## Usage
`nugoget update`

Automatically updates to the latest minor revision of dependencies. The sane
option to choose by default... looking at you `dotnet add package`


`nugoget update --major`

Automatically updates to the latest major revision of dependencies.
This will probably break your project.

`nugoget update --patch`

Automatically updates to the latest patch revision of dependencies.
This will probably fix your project. Unless you depend on bugs and you're _that_
person, then it'll break it.


`nugoget update --major --dryrun`

For when you want to see what will likely break during a major update.


`nugoget update -h`

For when you want to read the instructions.

## FAQ
- Why `go`?
  - I'm a go developer working in a dotnet world and I think the tooling is
    awful here
- Why... build this?
  - I have solutions with many projects within them, and updating the
    dependencies is extremely difficult from the command line. I find this to
    be a problem with a relatively easy solution.
- Why `go` and not `dotnet`?
  - BECAUSE I LIKE WRITING GO, STOP JUDGING ME.

## Contributing
Please don't. I don't have time. However, if you really want to, make it as easy
and clean as possible for me to merge. Maybe I'll find time.

### Contributors
Matthew General -- [https://digitalwny.com](digitalwny.com)

## License
[MIT License](./LICENSE)

