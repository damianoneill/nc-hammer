# NETCONF Load Tool

[![GitHub release](https://img.shields.io/github/release/damianoneill/nc-hammer.svg)](https://github.com/damianoneill/nc-hammer/releases)
[![Go Report Card](https://goreportcard.com/badge/damianoneill/nc-hammer)](http://goreportcard.com/report/damianoneill/nc-hammer)
[![license](https://img.shields.io/github/license/damianoneill/nc-hammer.svg)](https://github.com/damianoneill/nc-hammer/blob/master/LICENSE)

If you don't have a Go evnironment setup, you can __dowload a binary__ from the [releases](https://github.com/damianoneill/nc-hammer/releases) page, I suggest you place this somewhere in an existing bin path.

The tool uses a YAML file to define the test suite.  A sample [Test Suite](./suite/testdata/testsuite.yml) is included in the repository.  Running a scenario generates a results directory containing a copy of the testsuite definition and its results encoded in CSV format.  The tool can then be used to generate reports against the contents within the results directory.

![alt text](img/arch.png)

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

The blocks' configuration contains the defintion of the sequence of requests (an action) that should be executed against your SUT.  The blocks section contains a list of block definitions, __the list is executed sequentially per client__.  Each block section defines the type of block it is, options include; init, sequential or concurrent.  The blocks themselves contain a list of actions, currently two action types are supported; netconf and sleep.

A sleep Action is a pause in the execution of a block.  The sleep action defines a duration in Milliseconds.

A netconf Action is a definition for a NETCONF operation.  The NETCONF operations that are supported are [get](https://tools.ietf.org/html/rfc6241#page-48), [get-config](https://tools.ietf.org/html/rfc6241#page-35) and [edit-config](https://tools.ietf.org/html/rfc6241#page-37).  The parameters that are available for each netconf action reflect the parameters defined in the [NETCONF Specification](https://tools.ietf.org/html/rfc6241).  

For e.g. the NETCONF RPC message containing an edit-config operation

```xml
<rpc xmlns="urn:ietf:params:xml:ns:netconf:base:1.0" message-id="101">
   <edit-config>
      <target>
         <running />
      </target>
      <config>
         <top xmlns="http://example.com/schema/1.2/config">
            <interface>
               <name>Ethernet0/0</name>
               <mtu>1500</mtu>
            </interface>
         </top>
      </config>
   </edit-config>
</rpc>
```

maps to the following yaml definition in a testsuite

```yaml
- netconf:
    hostname: 10.0.0.1
    operation: edit-config
    target: running
    config: <top xmlns="http://example.com/schema/1.2/config"><protocols><ospf><area><name>0.0.0.0</name><interfaces><interface
      xc:operation="delete"><name>192.0.2.4</name></interface></interfaces></area></ospf></protocols></top>
```

Note that the config yaml tag contains the contents of the xml contained with the <config/> element within the rpc element

To reduce the verbosity of test suites the config attribute in the YAML file can be overriden to refer to an external XML file that should have its contents inlined using the __file:__ identifier.  For e.g. an XML file such as

```sh
$ cat edit-config.xml
<top xmlns="http://example.com/schema/1.2/config"><protocols><ospf><area><name>0.0.0.0</name><interfaces><interface operation="delete"><name>192.0.2.4</name></interface></interfaces></area></ospf></protocols></top>
```

can be referenced in the YAML as

```yaml
- netconf:
    hostname: 10.0.0.1
    operation: edit-config
    target: running
    config: file:edit-config.xml
```

When running the testsuite the content in the XML file will be inlined as if it had been defined as:

```yaml
- netconf:
    hostname: 10.0.0.1
    operation: edit-config
    target: running
    config: <top xmlns="http://example.com/schema/1.2/config"><protocols><ospf><area><name>0.0.0.0</name><interfaces><interface
      xc:operation="delete"><name>192.0.2.4</name></interface></interfaces></area></ospf></protocols></top>
```

#### Init

An init block is used to initialise the SUT, this is optional and is not required to execute a test suite.  If more than one init block is defined, the first one in the list is used.  The init block is executed once (regardless of number of clients or number of iterations), on suite startup before any other block is executed.

#### Sequential

A sequential block is a set of actions that are executed sequentially.  An assumption can be made with regard to ordering in this block type.

#### Concurrent

A concurrent block contains a set of actions that are executed concurrently.  No assumption should be made with regard to ordering in this block type.

### Handling XML

Some NETCONF Actions require defining snippets of XML for e.g. in the edit-config operation, any XML included in TestSuite should be minified, this can be simplified by using an [online minifier](http://www.webtoolkitonline.com/xml-minifier.html).

For e.g. 

```
<top xmlns="http://example.com/schema/1.2/config">
   <interface>
      <name>Ethernet0/0</name>
      <mtu>1500</mtu>
   </interface>
</top>
```

would become

```
<top xmlns="http://example.com/schema/1.2/config"><interface><name>Ethernet0/0</name><mtu>1500</mtu></interface></top>
```

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
  init        Scaffold a TestSuite and snippets directory
  run         Execute a Test Suite
  version     Show nc-hammer version

Flags:
      --config string   config file (default is $HOME/../nc-hammer.yaml)
  -h, --help            help for nc-hammer
  -t, --toggle          Help message for toggle

Use "nc-hammer [command] --help" for more information about a command.
```

## Example Usage

To simplify the process of getting up and running, the application includes a scaffolding function.  

```sh
nc-hammer init scenario1
```

This will generate a folder called scenario1 that includes a sample TestSuite and an example XML Snippet.

```sh
$ tree scenario1
scenario1
├── snippets
│   └── edit-config.xml
└── test-suite.yml
```

A Test Suite run can be executed as follows, note that as the suite runs, it will write a '.' to the screen to indicate a succesful NETCONF Request and a 'E' to indicate an Error.

```sh
$ nc-hammer run test-suite.yml
Testsuite /Users/doneill/scenario1/test-suite.yml started at Tue Jun 19 10:55:33 2018
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
$ nc-hammer analyse results/2018-06-19-10:55:55/

Testsuite executed at 2018-06-19-10:55:55
Suite defined the following hosts: [172.26.138.50 172.26.138.57 172.26.138.118 172.26.138.53 172.26.138.46]
5 client(s) started, 10 iterations per client, 0 seconds wait between starting each client

Total execution time: 22.368s, Suite execution contained 2 errors


 HOST           OPERATION   REUSE CONNECTION  REQUESTS  MEAN     VARIANCE   STD DEVIATION

 172.26.138.50  get-config  false                   48  2185.17  297421.42         545.36

```

As you can see the default analyse option generates the __mean__ (the total of the latencies divided by how many latencies there are), __variance__ (measures how far each latency in the set is from the mean) and __standard devitation__ (is a measure of the extent to which the latency set varies from the mean) for the set of latencies associated with a specific operation against a specific host.

If the results included errors (the latencies for these are excluded from the set of results), you can analyse the errors as follows:

```sh
$ nc-hammer analyse error results/2018-06-19-10:55:55/

Testsuite executed at 2018-06-19-10:55:55
Total Number of Errors for suite: 2

 HOSTNAME       OPERATION   ERROR

 172.26.138.50  get-config  ssh: handshake failed: ssh: unable to authenticate, attempted methods [none
                            password], no supported methods remain

 172.26.138.50  get-config  ssh: handshake failed: ssh: unable to authenticate, attempted methods [none
                            password], no supported methods remain
```

## Build

You should have a working go environment, packages are managed by [vgo](https://github.com/golang/go/wiki/vgo-user-guide).  This includes support for at least go v1.10.x.

```sh
go get -u golang.org/x/vgo
```

After cloning the repository, run vgo build to resolve imports and download the dependent packages.

```sh
vgo build
```

## Credits

The design is heavily influenced by [gotling](https://github.com/eriklupander/gotling), thanks to Erik Lupander, for the following article http://callistaenterprise.se/blogg/teknik/2015/11/22/gotling/ 