# Preparação de produção — evidências de 13/10/2026

## Integrado e implantado em homologação

PR 15 integrado em `93d1f5af0e26f588c2e21b8f039c4f5f4ff9c540`. CI de validação e segurança aprovadas tanto no PR quanto em main. Coolify fixado nessa revisão, mantendo deploy automático desligado.

Deploy `jgd7mryfxm1qy8p10uyzodmy`: iniciado às 17:00:36 -03, concluído com sucesso às 17:02:05 -03. Banco saudável; migração encerrou com código 0; API e web saudáveis. Nenhuma migração nova introduzida pelo PR 15. Imagens anteriores e volume preservados.

## Inspeção efetiva após atualização

- VPS `vmi3570746`, às 17:02:13 -03: 4998 MiB disponíveis; disco raiz com 76G livres. Medição pontual, não benchmark de pico nem garantia de capacidade.
- Memória: db 512 MiB, API 256 MiB, web 128 MiB, migrate 256 MiB.
- API/web/migrate com filesystem somente leitura; banco gravável com volume exclusivo `rk48swz6mdqduoyyhfnqhylp_postgres_data`.
- Banco e migrate somente na rede database; API em app/database; web em app/coolify. Rede database efetivamente interna.
- PortBindings de API, banco e web vazios: nenhuma porta publicada no host.
- Hash da lista ordenada de IDs/nomes dos demais containers: `6cc9a131cdb81e3964da7f2f1c6b623c2ba6a78d1cd2e41ba2843cd35a7cdbb6`, idêntico ao inventário anterior. Isso comprova preservação da identidade dos containers; não mede disponibilidade contínua de cada aplicação.

## Validação externa e autenticada

HTTPS confiável, página 200, health/ready 200, HTML no-cache e API no-store. Bundle servido `/assets/index-zJJ0vq_B.js` contém os novos painéis. Login realizado com os dados já preenchidos pelo responsável, sem ler ou registrar credenciais. Dois endpoints e histórico anterior preservados. Após recarregar a interface, Adicionar endpoint foca Nome dentro do viewport; Detalhes e histórico foca seu título e posiciona o conteúdo na tela. Formulário de teste cancelado, sem criar endpoint extra.

Check manual do endpoint já cadastrado pelo responsável (`franklem.com`) concluído às 17:04:08 -03: certificado válido, hostname e cadeia confiáveis; histórico passou de uma para duas amostras. Nenhum teste SSRF dirigido aos serviços reais. Logout e inspeção de atributos de cookie após este deploy não ensaiados nesta rodada; CI autenticada permanece aprovada.

## Configuração de produção preparada

`compose.coolify.prod.yaml` exclusivo, routers/service/middleware `certguard-prod`, origem e domínio `certguard.franklem.uk`, recurso/volume/segredo próprios. Nenhum recurso de produção criado. Scheduler, retenção, entregas e métricas continuam desabilitados.

Validação local com Docker Compose oficial v5.6.0, baixado somente para o diretório de ferramentas da tarefa e com SHA256 conferido. Quatro renderizações reais aprovadas: dois ambientes, nomes padrão e UUIDs simulados. Casos adulterados de origem incorreta e banco com porta pública rejeitados nos dois ambientes. O comando Docker instalado no host local é wrapper Podman/Compose 1 e não oferece saída JSON; não foi tratado como aprovação. CI ampliada para ambos os ambientes.

## Bloqueios restantes

- Backup externo: `franklin@frank-note2` resolve por Tailscale, mas SSH com acesso existente recusou autenticação por chave. Nenhuma credencial criada, segredo lido, rotina de backup alterada ou arquivo transferido. Solicitado acesso existente ao responsável. Backup e restauração real em destino descartável ainda não executados.
- Agendamento e alertas não homologados na VPS: permanecem desligados, sem envios externos. A ativação depende de backup e destinos/canais autorizados; não apresentar o ambiente como monitoramento automático ativo.
- Aceite completo e implantação de produção permanecem pendentes. Não usar banco de produção como destino de restauração. Não reverter schema ou executar limpeza global Docker.
