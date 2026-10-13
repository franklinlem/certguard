# Métricas e saúde operacional

`/health` é liveness. `/ready` verifica PostgreSQL e versões completas do schema esperado; 503 impede considerar a API pronta. A resposta não contém versão, SQL ou credenciais.

`METRICS_TOKEN` vazio desabilita `/metrics` (404). Use token aleatório ASCII de 32–256 caracteres no segredo do ambiente. A coleta usa `Authorization: Bearer <token>`, por acesso interno à API, não pelo domínio web. `/metrics` e `/api/metrics` no Nginx retornam 404. Não publicar porta da API na VPS para habilitar scrape; conectar coletor à rede isolada somente após inventário e autorização conforme risco.

Exemplo de configuração a adaptar no coletor existente, sem instalar serviço novo:

```yaml
scrape_configs:
  - job_name: certguard-dev
    scrape_interval: 30s
    scrape_timeout: 5s
    authorization:
      type: Bearer
      credentials_file: /run/secrets/certguard_metrics_token
    static_configs:
      - targets: ['API_INTERNA_CONFIRMADA:8080']
        labels:
          environment: homologacao
```

Target/rede são placeholders que dependem do inventário; não executar essa configuração sem substituir e revisar. Token pertence a um arquivo de segredo, nunca à URL. Coletar uma instância por banco/ambiente: valores duráveis são globais e réplicas exportam o mesmo total.

| Família | Tipo / significado |
|---|---|
| certguard_checks_total | Counter de checks persistidos |
| certguard_check_failures_total | Counter de checks com code != success |
| certguard_check_duration_seconds | Histograma cumulativo: bucket/sum/count |
| certguard_alert_deliveries_total | Counter de tentativas persistidas |
| certguard_alert_delivery_failures_total | Counter de tentativas sem aceite |
| certguard_alert_delivery_accepted_total | Counter de aceites persistidos |
| certguard_endpoints, certguard_endpoints_by_status | Gauges de não arquivados e último resultado |
| certguard_endpoints_enabled / paused / due | Gauges de configuração e atraso de agenda |
| certguard_certificates_expiring | Gauge de últimos certificados vencendo entre agora e 30 dias |
| certguard_alert_deliveries_pending / failed | Gauges de pendências e falhas terminais |
| certguard_alert_oldest_pending_age_seconds | Gauge da idade da pendência mais antiga |
| certguard_last_check_timestamp_seconds | Timestamp do último check persistido |
| certguard_last_delivery_attempt_timestamp_seconds | Timestamp da última tentativa persistida |

Contadores começam com registros ainda disponíveis na migração 006 e permanecem após retenção de checks. Não reconstruir histórico já removido. Contagem é lógica, baseada em commits; queda após envio e antes do commit pode repetir conexão sem refletir uma tentativa adicional persistida. Warning/critical por validade próxima não são falhas de inspeção quando code=success.

As 30 séries têm cardinalidade fixa; hostname/URL/e-mail/UUID não são etiquetas. Não há métricas por endpoint, HTTP, runtime ou worker heartbeat nesta entrega. Usar idade de checks/fila junto dos flags configurados; pause/flags desligados podem explicar valores altos. Definir limites de alerta após observar a homologação, sem inventar SLO de produção.
