FROM neo4j:latest

ENV NEO4J_AUTH=neo4j/test

ENV NEO4J_dbms_connector_https_advertised__address="localhost:7473"
ENV NEO4J_dbms_connector_http_advertised__address="localhost:7474"
ENV NEO4J_dbms_connector_bolt_advertised__address="localhost:7687"

VOLUME $HOME/neo4j/data:/data
VOLUME $HOME/neo4j/logs:/logs
VOLUME $HOME/neo4j/import:/var/lib/neo4j/import
VOLUME $HOME/neo4j/plugins:/plugins