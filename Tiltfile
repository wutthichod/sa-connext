# ============================================================================
# Docker Builds
# ============================================================================

docker_build(
    'connext-api-gateway:latest',
    '.',
    dockerfile='infra/dockerfile/api-gateway',
    build_args={'SERVICE_NAME': 'api-gateway'},
    only=['shared/', 'services/api-gateway/', 'go.mod', 'go.sum']
)

docker_build(
    'connext-user-service:latest',
    '.',
    dockerfile='infra/dockerfile/user-service',
    build_args={'SERVICE_NAME': 'user-service'},
    only=['shared/', 'services/user-service/', 'go.mod', 'go.sum']
)

docker_build(
    'connext-chat-service:latest',
    '.',
    dockerfile='infra/dockerfile/chat-service',
    build_args={'SERVICE_NAME': 'chat-service'},
    only=['shared/', 'services/chat-service/', 'go.mod', 'go.sum']
)

docker_build(
    'connext-event-service:latest',
    '.',
    dockerfile='infra/dockerfile/event-service',
    build_args={'SERVICE_NAME': 'event-service'},
    only=['shared/', 'services/event-service/', 'go.mod', 'go.sum']
)

docker_build(
    'connext-notification-service:latest',
    '.',
    dockerfile='infra/dockerfile/notification-service',
    build_args={'SERVICE_NAME': 'notification-service'},
    only=['shared/', 'services/notification-service/', 'go.mod', 'go.sum']
)

docker_build(
    'connext-mongo:latest',
    '.',
    dockerfile='infra/dockerfile/mongo',
    only=['infra/mongo/']
)

# ============================================================================
# Kubernetes Manifests
# k8s_yaml() automatically watches these files for changes
# ============================================================================

k8s_yaml([
    # Infrastructure
    'infra/k8s/app-config.yaml',
    'infra/k8s/app-secrets.yaml',
    'infra/k8s/db-secrets.yaml',
    'infra/k8s/postgres-db.yaml',
    'infra/k8s/pgadmin-db.yaml',
    'infra/k8s/messaging-and-mongo.yaml',
    'infra/k8s/rabbitmq.yaml',
    
    # Services - Deployments
    'infra/k8s/user-service-deployment.yaml',
    'infra/k8s/chat-service-deployment.yaml',
    'infra/k8s/event-service-deployment.yaml',
    'infra/k8s/notification-service-deployment.yaml',
    
    # Services - Services
    'infra/k8s/user-service.yaml',
    'infra/k8s/chat-service.yaml',
    'infra/k8s/event-service.yaml',
    
    # Gateway
    'infra/k8s/api-gateway-deployment.yaml',
    'infra/k8s/api-gateway-service.yaml',
])

# ============================================================================
# Infrastructure Resources
# ============================================================================

k8s_resource(
    objects=['mongodb-secret', 'pgadmin-secret', 'postgres-secret'],
    new_name='Database Secrets',
    labels='Infrastructure'
)

k8s_resource(
    objects=['app-secret', 'app-config'],
    new_name='App Config & Secrets',
    labels='Infrastructure'
)

k8s_resource(
    objects=['postgres-pvc'],
    new_name='Postgres Volume',
    labels='Database'
)

k8s_resource(
    workload='postgres-deployment',
    new_name='Postgres',
    resource_deps=['Database Secrets', 'Postgres Volume'],
    labels='Database',
    port_forwards=5433
)

k8s_resource(
    workload='pgadmin-deployment',
    new_name='PgAdmin',
    resource_deps=['Postgres'],
    port_forwards=5051,
    labels='Database'
)

k8s_resource(
    objects=['mongo-pvc'],
    new_name='Mongo Volume',
    labels='Database'
)

k8s_resource(
    workload='mongodb-deployment',
    new_name='MongoDB',
    resource_deps=['Database Secrets', 'Mongo Volume'],
    labels='Database',
    pod_readiness='wait'
)

# ============================================================================
# RabbitMQ Message Broker
# ============================================================================

k8s_resource(
    objects=['rabbitmq-pvc'],
    new_name='RabbitMQ Volume',
    labels='MessageBroker'
)

k8s_resource(
    workload='rabbitmq-deployment',
    new_name='RabbitMQ',
    resource_deps=['RabbitMQ Volume'],
    labels='MessageBroker',
    port_forwards='15672:15672'
)

# ============================================================================
# Application Services
# Wait for databases and RabbitMQ to be ready before starting services
# ============================================================================

k8s_resource(
    workload='user-service-deployment',
    new_name='User Service',
    resource_deps=['Postgres', 'RabbitMQ', 'App Config & Secrets'],
    labels='Service',
    pod_readiness='wait'
)

k8s_resource(
    workload='chat-service-deployment',
    new_name='Chat Service',
    resource_deps=['MongoDB', 'RabbitMQ', 'App Config & Secrets'],
    labels='Service',
    pod_readiness='wait'
)

k8s_resource(
    workload='event-service-deployment',
    new_name='Event Service',
    resource_deps=['Postgres', 'RabbitMQ', 'App Config & Secrets'],
    labels='Service',
    pod_readiness='wait'
)

k8s_resource(
    workload='notification-service-deployment',
    new_name='Notification Service',
    resource_deps=['RabbitMQ', 'App Config & Secrets'],
    labels='Service'
)

# ============================================================================
# API Gateway
# ============================================================================

k8s_resource(
    workload='api-gateway-deployment',
    new_name='API Gateway',
    resource_deps=['User Service', 'Chat Service', 'Event Service', 'RabbitMQ', 'App Config & Secrets'],
    port_forwards=8080,
    labels='Gateway',
    pod_readiness='wait'
)

# Tilt exposes localhost:8080. kind-config.yaml also maps NodePort 30080 to
# host port 8080; other devices can use http://<YOUR_IP>:8080 with that mapping.
