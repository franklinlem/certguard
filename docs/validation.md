# Validação da etapa inicial

Data: 12/10/2026. Código validado: `9006d64e2a017213a5f7cd31be7be5548624a3be`.

## Verificações locais
- Go 1.27.2: build, go vet e Staticcheck v0.8.2 aprovados.
- Testes Go com detector de condições de corrida aprovados.
- Cobertura dos pacotes: certificate 91,8%; security 97,8%; httpapi 100%.
- Fuzzing do parser: 64.421 entradas em cerca de 11 segundos, sem falhas.
- Certificados efêmeros e servidores TLS locais; nenhum domínio externo foi varrido.
- Frontend: lint, TypeScript, teste de componente e build aprovados; página conferida no navegador.
- npm audit e Govulncheck v1.9.0 aprovados. x/text foi atualizado para v0.41.0 após a auditoria identificar vulnerabilidades transitivas.

## GitHub Actions
- [Validação](https://github.com/franklinlem/certguard/actions/runs/38291231199): backend, frontend e Docker.
- [Segurança](https://github.com/franklinlem/certguard/actions/runs/38291231209): dependências Go, scanner de código/segredos e imagens.
- PostgreSQL 17 em runner efêmero: migração idempotente, constraints e gravação/leitura de histórico.
- Docker Compose: build, inicialização do zero, migrações, saúde/readiness e proxy web→API aprovados.
- Auditoria Trivy: nenhum achado bloqueante HIGH/CRITICAL no código; imagens sem achados corrigíveis HIGH/CRITICAL. Vulnerabilidades sem correção nas imagens são excluídas desse gate; o resultado não afirma ausência absoluta de vulnerabilidades.

## Limites
Docker e PostgreSQL não estão instalados no computador local; as evidências desses componentes vêm da CI. O teste PostgreSQL é pulado localmente quando TEST_DATABASE_URL não existe, mas executado na CI.

Autenticação, CRUD, persistência dos resultados do motor, scheduler, alertas e dashboard funcional ainda não implementados. A fundação não equivale ao MVP concluído.

Não há evidência de deploy: nenhum serviço da VPS foi acessado ou alterado, e nenhum ambiente de homologação/produção foi publicado. Proteção de main e CodeQL privado devem ser configurados conforme permissões/plano GitHub antes do fluxo colaborativo; não foram habilitados nesta etapa.
