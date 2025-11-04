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
    
    # Services
    'infra/k8s/user-service-deployment.yaml',
    'infra/k8s/chat-service-deployment.yaml',
    'infra/k8s/event-service-deployment.yaml',
    'infra/k8s/notification-service-deployment.yaml',
    
    # Gateway
    'infra/k8s/api-gateway-service-deployment.yaml',
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
    labels='Database'
)

k8s_resource(
    workload='pgadmin-deployment',
    new_name='PgAdmin',
    resource_deps=['Postgres'],
    port_forwards=5051,
    labels='Database'
)

# ============================================================================
# Application Services
# ============================================================================

k8s_resource(
    workload='user-service-deployment',
    new_name='User Service',
    resource_deps=['Postgres', 'App Config & Secrets'],
    labels='Service'
)

k8s_resource(
    workload='chat-service-deployment',
    new_name='Chat Service',
    resource_deps=['Postgres', 'App Config & Secrets'],
    labels='Service'
)

k8s_resource(
    workload='event-service-deployment',
    new_name='Event Service',
    resource_deps=['Postgres', 'App Config & Secrets'],
    labels='Service'
)

k8s_resource(
    workload='notification-service-deployment',
    new_name='Notification Service',
    resource_deps=['App Config & Secrets'],
    labels='Service'
)

# ============================================================================
# API Gateway
# ============================================================================

k8s_resource(
    workload='api-gateway-deployment',
    new_name='API Gateway',
    resource_deps=[
        'User Service',
        'Chat Service',
        'Event Service',
        'Notification Service',
        'App Config & Secrets'
    ],
    port_forwards=8080,
    labels='Gateway'
)
