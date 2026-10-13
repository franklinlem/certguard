# Preparação para produção — 13/10/2026

## Escopo autorizado

Concluir pendências de homologação e preparar produção. Preservar outros serviços e dados. Publicação de produção depende do aceite de homologação; operações com risco de interrupção/perda exigem autorização específica. Não criar credenciais persistentes nem alterar SSH para contornar acesso indisponível.

## Sequência

1. Conferir CI e integrar PR 15; fixar o commit e atualizar somente CertGuard-dev pelo Coolify. Confirmar readiness, versão web e fluxo de navegação com a sessão disponível.
2. Inspecionar somente saúde, redes, portas, limites e recursos da VPS; não imprimir ambientes ou segredos. Comparar serviços existentes e verificar isolamento real.
3. Inventariar frank-note2 via acesso existente. Definir diretório exclusivo, armazenamento criptografado, política de retenção e agendamento sem tocar em outros backups. Fazer backup do banco CertGuard e restaurar em banco/instância descartável nova, sem sobrescrever dados. Não configurar mecanismo de acesso ou criptografia sem conhecer a estrutura existente.
4. Versionar Compose exclusivo de produção com domínio, nomes Traefik e identidade próprios; ampliar verificações de política na CI para ambos os ambientes. Manter scheduler, retenção e entregas desativados na implantação inicial.
5. Registrar pendências concretas de backup/acesso, validação de scheduler/alertas com destinos autorizados, recuperação administrativa e rollback. Publicar configuração em PR e aguardar CI.
6. Criar recurso de produção separado somente após gates satisfeitos; verificar HTTPS, readiness, ausência de exposição do banco/API e proteção de autenticação. Administrador e segredos definidos por entrada segura, sem senha na conversa. Não apresentar implantação manual como monitoramento automático completo.

## Aceite e rollback

CI verde, correções validadas em homologação, recursos efetivos isolados, backup externo verificado e restauração descartável aprovada. Preservar volume e commit anterior em rollback da aplicação; não reverter schema, apagar dados, limpar Docker ou modificar proxy global. Cada evidência deve distinguir leitura, configuração e execução comprovada.
