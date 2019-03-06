# FileStack Tool
## Installation
Download the binary file:
```
curl -L https://github.com/gr00by87/fst/raw/master/bin/fst -o /usr/local/bin/fst && curl -L https://github.com/gr00by87/fst/raw/master/bin/fst-core -o /usr/local/bin/fst-core && chmod 755 /usr/local/bin/fst /usr/local/bin/fst-core
```

Run config and provide AWS IAM user security credentials with `ec2:DescribeInstances` permission:
```
fst config
```

You're all set!

## Usage
Check out `--help` for detailed usage
```
fst --help
```
