docker run -d --name service-registry \
    --network rtcheap -p 8080:8080 \
    -e MIGRATIONS_PATH='/etc/service-registry/migrations' \
    -e JWT_ISSUER='rtcheap' \
    -e JWT_SECRET='password' \
    -e DB_HOST='rtcheap-db' \
    -e DB_PORT='3306' \
    -e DB_DATABASE='serviceregistry' \
    -e DB_USERNAME='serviceregistry' \
    -e DB_PASSWORD='password' \
    -e JAEGER_SERVICE_NAME='service-registry' \
    -e JAEGER_SAMPLER_TYPE='const' \
    -e JAEGER_SAMPLER_PARAM=1 \
    -e JAEGER_REPORTER_LOG_SPANS='1' \
    -e JAEGER_AGENT_HOST='jaeger' \
    eu.gcr.io/rtcheap/service-registry:0.1.0