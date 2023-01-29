# ibm mq
installation xk6

go install go.k6.io/xk6/cmd/xk6@latest
### for linux 
mkdir /opt/mqm
cd /opt/mqm
tar -xvf 9.3.1.0-IBM-MQC-Redist-LinuxX64.tar.gz

export MQ_INSTALLATION_PATH=/opt/mqm  
export CGO_CFLAGS="-I$MQ_INSTALLATION_PATH/inc"
export CGO_LDFLAGS="-L$MQ_INSTALLATION_PATH/lib64 -Wl,-rpath,$MQ_INSTALLATION_PATH/lib64"

## Assembly
when assembling directory should contain request.100kb.xml with request body

CGO_ENABLED=1 xk6 build --output k6mq --with k6ibmmq=.

or

CGO_ENABLED=1 xk6 build --output k6mq --with github.com/ChipArtem/k6ibmmq
## Tests

example test.js.Example

password and login are set by environment variables IBMMQ_USER IBMMQ_PASS or a function in the test

newconn.setcredentials('login', 'password')
## Run

./k6mq run test.js