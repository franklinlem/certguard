# Plano de execução — Fundação e TLS Check Engine

Data: 12/10/2026. Referência: planejamento completo fornecido pelo proprietário.

## Objetivo e limites
Entregar monorepositório privado franklinlem/certguard, fundação reproduzível e motor TLS seguro testado. Esta etapa não é o MVP completo. Nenhuma conexão administrativa, alteração ou implantação na VPS está autorizada nesta execução. Produção depende de homologação validada.

## Sequência
1. Registrar planejamento original e decisões arquiteturais.
2. Implementar normalização de hostname, política de porta 443, DNS e bloqueio de todos os endereços especiais/privados.
3. Fixar IP validado na conexão, manter hostname em SNI e validar cadeia/hostname separadamente; coletar metadados e classificar falhas.
4. Testar com certificados efêmeros e servidores TLS isolados: validade, expiração, validade futura, SANs, wildcard, CA desconhecida, hostname incorreto, DNS misto/rebinding e timeout.
5. Preparar API de saúde, PostgreSQL e migração inicial; CLI local de inspeção. Não expor verificação arbitrária por HTTP sem autenticação.
6. Após estabilizar motor, preparar frontend React/TypeScript mínimo, sem simular dashboard funcional.
7. Preparar Docker Compose isolado, CI e instruções Coolify/Traefik para ambientes distintos. Sem deploy automático.
8. Executar testes, lint, build, cobertura core >=80%; registrar limitações reais do ambiente.
9. Criar repositório privado e publicar commits se acesso disponível; registrar bloqueio concreto caso autenticação não permita criação.

## Aceite
- Motor sem HTTP fetch, proxy ou segunda resolução DNS no dial.
- Endereços proibidos nunca alcançam o conector; configuração de teste não exposta ao usuário.
- Handshake não implica confiança; validação x509 explícita obrigatória.
- Timeout total cobre DNS/TCP/TLS; conexão sempre fechada.
- Resultados estruturados, fingerprint SHA-256, cadeia e SANs, dias e limiares.
- Segredos fora do Git; bancos, volumes e projetos Compose isolados por ambiente.

## Incrementos seguintes
Autenticação e CRUD, persistência de verificações, scheduler, dashboard, alertas, métricas e hardening. Implantação em homologação após revisão da infraestrutura existente; autorização específica antes de mudanças com risco de indisponibilidade/perda de dados.
