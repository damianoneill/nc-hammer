# NETCONF Load Tool

[![GitHub release](https://img.shields.io/github/release/damianoneill/nc-hammer.svg)](https://github.com/damianoneill/nc-hammer/releases)
[![Go Report Card](https://goreportcard.com/badge/damianoneill/nc-hammer)](http://goreportcard.com/report/damianoneill/nc-hammer)
[![license](https://img.shields.io/github/license/damianoneill/nc-hammer.svg)](https://github.com/damianoneill/nc-hammer/blob/master/LICENSE)

If you don't have a Go evnironment setup, you can __dowload a binary__ from the [releases](https://github.com/damianoneill/nc-hammer/releases) page.

The tool uses a yaml file to define the test setup.  A sample [Test Suite](./suite/testdata/testsuite.yml) is included in the repository.

## Test Suite

A Test Suite is a YAML document that is used to feed nc-hammer.  It is made up of three sections; Suite Configuration, Host Configuration and a section that contains the sequences of Actions (primarily netconf requests) to be executed.

__YAML is sensitive to indentation__, if your not familiar with YAML see [here](https://learnxinyminutes.com/docs/yaml/).

### Suite Configuration

The suite configuration defines the top-level setup for a Test Suite, this includes configuration options for;

* The number of iterations that the block section should be repeated for
* The number of concurrent clients that should connect to each Host
* A rampup time for the client connections

These permutations allow you to do both functional (iterations:1 and concurrent:1) and load (concurrent:n, where n>1) testing.

### Host Configuration

The host configuration defines the parameters required to make a SSH connection to a Device.  This includes;

* hostname (dns or ip name)
* port (for netconf agents running on a nonstandard port)
* username (netconf username)
* password (netconf password)
* reuseconnection (indicates whether a ssh connection against a device should be reused or restablished each time a request is sent)

### Blocks Configuration

The blocks' configuration contains the defintion of the sequence of requests (an action) that should be executed against your SUT.  The blocks section contains a list of block definitions, __the list is executed sequentially__.  Each block section defines the type of block it is, options include; init, sequential or concurrent.  The blocks themselves contain a list of actions, currently two action types are supported; netconf and sleep.

#### Init

An init block is used to initialise the SUT, this is optional and is not required to execute a test suite.  If more than one init block is defined, the first one in the list is used.  The init block is executed once (regardless of number of clients or number of iterations), on suite startup before any other block is executed.

#### Sequential

A sequential block is a set of actions that are executed sequentially.  An assumption can be made with regard to ordering in this block type.

#### Concurrent

A concurrent block contains a set of actions that are executed concurrently.  No assumption should be made with regard to ordering in this block type.

## Usage

```sh
$ nc-hammer
A NETCONF Load Tester

Usage:
  nc-hammer [command]

Available Commands:
  analyse     Analyse the output of a Test Suite run
  completion  Generate shell completion script for nc-hammer
  help        Help about any command
  run         Execute a Test Suite
  version     Show ./nc-hammer version

Flags:
      --config string   config file (default is $HOME/../nc-hammer.yaml)
  -h, --help            help for nc-hammer
  -t, --toggle          Help message for toggle

Use "nc-hammer [command] --help" for more information about a command.
```

## Example Usage

A Test Suite run can be executed as follows, note that as the suite runs, it will write a '.' to the screen to indicate a succesful NETCONF Request and a 'E' to indicate an Error.

```sh
$ ./nc-hammer run ~/test-suite.yml
Testsuite /Users/doneill/test-suite.yml started at Tue Jun 19 10:55:33 2018
 > 5 client(s) started, 10 iterations per client, 0 seconds wait between starting each client
.................E................E...............
Testsuite completed in 22.369465719s
```

After completion of a testsuite, an output folder with the date timestamp will be created in a folder called results.

```sh
$ ls results
2018-06-19-10:55:55
```

You can analyse the results as follows:

```sh
$ ./nc-hammer analyse results/2018-06-19-10:55:55/

Testsuite executed at 2018-06-19-10:55:55
Suite defined the following hosts: [172.26.138.50 172.26.138.57 172.26.138.118 172.26.138.53 172.26.138.46]
5 client(s) started, 10 iterations per client, 0 seconds wait between starting each client

Total execution time: 22.368s, Suite execution contained 2 errors


 HOST           OPERATION   REUSE CONNECTION  REQUESTS  MEAN     VARIANCE   STD DEVIATION

 172.26.138.50  get-config  false                   48  2185.17  297421.42         545.36

```

If the results included errors, you can analyse the errors as follows:

```sh
$ ./nc-hammer analyse error results/2018-06-19-10:55:55/

Testsuite executed at 2018-06-19-10:55:55
Total Number of Errors for suite: 2

 HOSTNAME       OPERATION   ERROR

 172.26.138.50  get-config  ssh: handshake failed: ssh: unable to authenticate, attempted methods [none
                            password], no supported methods remain

 172.26.138.50  get-config  ssh: handshake failed: ssh: unable to authenticate, attempted methods [none
                            password], no supported methods remain
```

## Build

You should have a working go environment, packages are managed by [vgo](https://github.com/golang/go/wiki/vgo-user-guide).

```sh
go get -u golang.org/x/vgo
```

After cloning the repository, run vgo build to resolve imports and download the dependent packages.

```sh
vgo build
```

## Credits

The design is heavily influenced by [gotling](https://github.com/eriklupander/gotling), thanks to Erik Lupander, for the following article http://callistaenterprise.se/blogg/teknik/2015/11/22/gotling/ 