# Segurança — etapa 07

SSRF e inspeção estão nos ADR-002/003. Somente porta 443 e IPs aprovados; DNS fixado no dial. Nenhum bypass por ambiente. Certificados e seus metadados são dados não confiáveis: JSON e renderização como texto, sem HTML recebido do alvo.

Acesso administrativo com bcrypt custo 12, bootstrap por stdin sem senha padrão, sessões de oito horas e token de 256 bits armazenado como hash. Cookie host-only, HttpOnly, SameSite=Strict e Secure fora do ambiente local. Logout revoga a sessão atual.

Mutações exigem Origin exata e token CSRF; login exige Origin e limites de tentativas/concorrência. Corpos JSON até 8 KiB. Nenhum segredo em logs, argumentos da CLI ou Git. O ADR-004 registra limitações de rate limiting em memória e do IP compartilhado atrás de proxy; limite distribuído fica para evolução da infraestrutura.

Checks manuais exigem sessão, no máximo quatro execuções por processo e uma por endpoint, com cooldown de 30 segundos. Cadastro valida sintaxe; DNS/IP são revalidados em toda inspeção. Alterações e arquivamento não apagam o histórico.

Docker API/web sem root, filesystem somente leitura, capabilities removidas e no-new-privileges. Banco em rede interna, sem portas públicas, volume e secrets exclusivos por ambiente. A imagem PostgreSQL inicializa o volume e reduz privilégios conforme seu padrão; não aplicar read-only ao diretório de dados.

Para Compose use senha PostgreSQL alfanumérica: caracteres reservados exigem codificação na URL. Cabeçalhos de segurança e noindex no frontend; HSTS pertence ao proxy após validar HTTPS, sem impor includeSubDomains ao domínio compartilhado.

Antes de homologar: rever rotas/egress, recursos, backups/restore, proteção de main e secrets. Nenhuma alteração na VPS sem avaliar impacto; alterações com risco de interrupção/perda de dados exigem autorização. CodeQL privado depende do plano GitHub e não foi habilitado automaticamente.

O agendador inspeciona somente endpoints ativos cadastrados pelo administrador. Reutiliza integralmente o motor SSRF/TLS, sem bypass. Até 25 workers configuráveis por instância (padrão 10), além dos quatro checks manuais; locks no PostgreSQL impedem sobreposição por endpoint entre instâncias. Retenção é opt-in, limitada a 1000 registros por lote e janela mínima de 30 dias, mantendo o último resultado. Não há garantia de exatamente uma conexão externa se o processo cair antes do commit.

Eventos são consultáveis apenas com sessão administrativa. Regras não iniciam conexões; o executor separado entrega somente em canais ativos autorizados; observações e destinos não confiáveis são renderizados como texto. Eventos persistem junto do check, sob lock, sem parâmetros SQL interpolados. Retenção dos checks apenas anula a referência e mantém o snapshot. Autorização para implantação e transporte externo continua independente desta entrega.

Configurações de canais cifradas com AES-256-GCM e AAD. Chave fora do banco; validação na ativação/partida do executor. Cadastro não envia teste e começa pausado. SMTP exige TLS antes de autenticação; webhook valida TLS, DNS público integral, IP fixo, sem redirects/proxy e HMAC. Erros remotos sanitizados. API devolve apenas metadados. Veja ADR-008 e operação para backup, backlog e limites de idempotência.

Métricas opt-in com token separado, comparação constante e deadline; resposta sem identificação/segredos. Nginx encaminha apenas /api/v1, /api/health e /api/ready, bloqueando /metrics e demais /api. Readiness rejeita versões ausentes/futuras. Schema e índices aditivos não dispensam revisão de compatibilidade/rollback. Veja roteiro de homologação.
