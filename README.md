# Kmind Agent Docker

Agente Docker do Kmind. Esta primeira fundacao valida configuracao, identidade persistente, conectividade segura com o Docker Engine e controle de ativacao. A descoberta de containers, metricas, eventos e logs sera adicionada nas proximas fases.

O agente envia somente OTLP/HTTP Protobuf comprimido ao Gateway Kmind. Nunca conecta diretamente ao Loki, Tempo ou VictoriaMetrics e nunca executa operacoes mutaveis na Docker Engine.

## Chave e identidade

Cada integracao Docker recebe uma chave privada `kmd_...`, armazenada pelo Kmind somente como hash. Uma mesma chave pode ser usada em varios hosts que pertencem ao mesmo servico Docker. O volume `/var/lib/kmind-agent` guarda somente o identificador aleatorio persistente da instalacao.

## Configuracao

Copie `configs/config.example.yaml`, informe a chave por segredo/arquivo e mantenha o arquivo inacessivel a outros usuarios. Variaveis de ambiente tem precedencia sobre YAML, que tem precedencia sobre defaults internos.

## Estado atual

Fase 1 em implementacao. Nao use este projeto ainda para producao.
