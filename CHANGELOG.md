<a name="unreleased"></a>
## [Unreleased]


<a name="1.4.1"></a>
## [1.4.1] - 2020-04-05
### Refactor
- **vpn:** Send exactly the same payload to `/profile` handler as `Pritunl` would


<a name="1.4.0"></a>
## [1.4.0] - 2020-04-04
### Build
- **modules:** Add go modules

### Feat
- **vpn:** Add `vpn` subcommannd

### Fix
- **core:** List servers nested under `res.Instances`

### Refactor
- **scp:** Don't use bastion host in `us-east-1` region
- **ssh:** Don't use bastion host in `us-east-1` region


<a name="1.3.0"></a>
## [1.3.0] - 2019-03-25
### Feat
- **ssh:** Added support for `-N` ssh flag


<a name="1.2.1"></a>
## [1.2.1] - 2019-03-22
### Fix
- **list-servers:** Fix flag parsing for `list-servers` command


<a name="1.2.0"></a>
## [1.2.0] - 2019-03-18
### Feat
- **list-servers:** Add `ls` subcommand (alias of `list-servers`)
- **scp:** Add `scp` subcommand
- **ssh:** Add support for `-L` and `-F` ssh flags
- **ssh-config:** Add `ssh-config` subcommand


<a name="1.1.0"></a>
## [1.1.0] - 2019-03-06
### Feat
- **ssh:** Add ssh subcommand

### Refactor
- **filters:** Refactor filters to be more universal
- **list-servers:** Add `PRIVATE IP` column


<a name="1.0.1"></a>
## [1.0.1] - 2019-03-04
### Refactor
- **flags:** Use `cobra` instead of `go-flags` for flag parsing


<a name="1.0.0"></a>
## 1.0.0 - 2019-03-03
### Feat
- **app:** First release


[Unreleased]: https://github.com/gr00by87/fst/compare/1.4.1...HEAD
[1.4.1]: https://github.com/gr00by87/fst/compare/1.4.0...1.4.1
[1.4.0]: https://github.com/gr00by87/fst/compare/1.3.0...1.4.0
[1.3.0]: https://github.com/gr00by87/fst/compare/1.2.1...1.3.0
[1.2.1]: https://github.com/gr00by87/fst/compare/1.2.0...1.2.1
[1.2.0]: https://github.com/gr00by87/fst/compare/1.1.0...1.2.0
[1.1.0]: https://github.com/gr00by87/fst/compare/1.0.1...1.1.0
[1.0.1]: https://github.com/gr00by87/fst/compare/1.0.0...1.0.1
