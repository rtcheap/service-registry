cd cmd
go build
cd ..
mv cmd/cmd service-registry

export MIGRATIONS_PATH='./resources/db/mysql'
export JWT_ISSUER='rtcheap'
export JWT_SECRET='password'

export DB_HOST='127.0.0.1'
export DB_PORT='3306'
export DB_DATABASE='serviceregistry'
export DB_USERNAME='serviceregistry'
export DB_PASSWORD='password'

export JAEGER_SERVICE_NAME='service-registry'
export JAEGER_SAMPLER_TYPE='const'
export JAEGER_SAMPLER_PARAM=1
export JAEGER_REPORTER_LOG_SPANS='1'

./service-registry