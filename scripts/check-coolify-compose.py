"""Valida a política de isolamento dos ambientes no JSON renderizado pelo Docker Compose."""
import json
import sys

config = json.load(sys.stdin)
expected_project = sys.argv[1] if len(sys.argv) > 1 else 'certguard-dev'
environment = sys.argv[2] if len(sys.argv) > 2 else 'homologacao'
assert environment in {'homologacao', 'producao'}, 'ambiente desconhecido'
router = 'certguard-prod' if environment == 'producao' else 'certguard-dev'
domain = 'certguard.franklem.uk' if environment == 'producao' else 'certguard-dev.franklem.uk'
assert config['name'] == expected_project, 'identidade Compose divergente'
assert config['networks']['database']['name'] == expected_project + '_database'
assert config['networks']['app']['name'] == expected_project + '_app'
assert config['volumes']['postgres_data']['name'] == expected_project + '_postgres_data'
services = config['services']
assert set(services) == {'db', 'migrate', 'api', 'web'}
expected_networks = {
    'db': {'database'}, 'migrate': {'database'},
    'api': {'database', 'app'}, 'web': {'app', 'proxy'},
}
for name, service in services.items():
    assert not service.get('ports'), f'{name}: porta publicada no host'
    assert not service.get('privileged'), name
    assert not service.get('network_mode'), name
    assert set(service['networks']) == expected_networks[name], name
    if name != 'web':
        assert not any(k.startswith('traefik.') for k in service.get('labels', {})), name
    if name != 'db':
        assert service['read_only'], name
        assert service['cap_drop'] == ['ALL'], name
    assert int(service['mem_limit']) <= 512 * 1024 * 1024, name
    for volume in service.get('volumes', []):
        assert volume['type'] == 'volume', 'bind mount inesperado'
assert config['networks']['database']['internal']
assert config['networks']['proxy']['external']
assert config['networks']['proxy']['name'] == 'coolify'
assert not config['networks']['app'].get('external')
assert not config['volumes']['postgres_data'].get('external')
env = services['api']['environment']
assert env['APP_ENV'] == 'production'
assert env['PUBLIC_APP_URL'] == 'https://' + domain
for flag in ['SCHEDULER_ENABLED', 'RETENTION_ENABLED', 'DELIVERY_ENABLED']:
    assert env[flag] == 'false', flag
assert env['METRICS_TOKEN'] == env['ALERT_ENCRYPTION_KEY'] == ''
assert 'DATABASE_URL' not in services['web'].get('environment', {})
labels = services['web']['labels']
assert labels['traefik.docker.network'] == 'coolify'
assert labels[f'traefik.http.routers.{router}.rule'] == f'Host(`{domain}`)'
assert labels[f'traefik.http.routers.{router}.tls.certresolver'] == 'letsencrypt'
assert labels[f'traefik.http.services.{router}.loadbalancer.server.port'] == '8080'
assert services['api']['depends_on']['migrate']['condition'] == 'service_completed_successfully'
assert services['migrate']['depends_on']['db']['condition'] == 'service_healthy'
other_router = 'certguard-dev' if environment == 'producao' else 'certguard-prod'
assert not any(other_router in key for key in labels), 'rota de outro ambiente'
assert labels[f'traefik.http.routers.{router}-http.rule'] == f'Host(`{domain}`)'
assert labels[f'traefik.http.routers.{router}.service'] == router
assert labels[f'traefik.http.routers.{router}.entrypoints'] == 'https'
assert labels[f'traefik.http.routers.{router}.tls'] == 'true'
assert labels[f'traefik.http.routers.{router}-http.entrypoints'] == 'http'
assert labels[f'traefik.http.routers.{router}-http.middlewares'] == router + '-https'
assert labels[f'traefik.http.middlewares.{router}-https.redirectscheme.scheme'] == 'https'
print(f'Política de isolamento de {environment} aprovada.')
