# Tiltfile for sa-connext (fixed)

# -- Build images for services -------------------------------------------------
docker_build('connext-api-gateway:local', 'services/api-gateway', dockerfile='services/api-gateway/dockerfile')
docker_build('connext-user-service:local', 'services/user-service', dockerfile='services/user-service/dockerfile')
docker_build('connext-chat-service:local', 'services/chat-service', dockerfile='services/chat-service/dockerfile')
docker_build('connext-event-service:local', 'services/event-service', dockerfile='services/event-service/dockerfile')
docker_build('connext-notification-service:local', 'services/notification-service', dockerfile='services/notification-service/dockerfile')

# -- Apply k8s YAMLs -----------------------------------------------------------
# Put all YAML files here (Tilt will apply them in order)
k8s_yaml([
  'k8s/app-config.yaml',
  'k8s/app-secret.yaml',
  'k8s/db-secrets.yaml',
  'k8s/postgres-db.yaml',
  'k8s/pgadmin-db.yaml',

  'k8s/api-gateway-deployment.yaml',
  'k8s/api-gateway-service.yaml',

  'k8s/user-service-deployment.yaml',
  'k8s/user-service.yaml',

  'k8s/chat-service-deployment.yaml',
  'k8s/chat-service.yaml',

  'k8s/event-service-deployment.yaml',
  'k8s/event-service.yaml',

  'k8s/notification-service-deployment.yaml',
])

# -- Register resources for Tilt UI (optional port-forwards) ------------------
# Names here should match the metadata.name of your Deployments or Services
k8s_resource('api-gateway-deployment', port_forwards=8080)
k8s_resource('user-service-deployment', port_forwards=8081)
k8s_resource('chat-service-deployment', port_forwards=8082)
k8s_resource('event-service-deployment', port_forwards=8084)

# If you want to add resource dependencies or custom behavior, you can use
# k8s_resource(..., resource_deps=['other-resource-name']) but avoid passing
# unknown keyword args like yaml_files which Tilt's k8s_resource doesn't accept.