package managers

import (
	"errors"
	"fmt"
	"os"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

const relationshipIdentifier = "CONNECTED"

var neo4jHost = os.Getenv("NEO4J_HOST")

type Neo4jManager struct {
	driver neo4j.Driver
}

func NewNeo4jManager() *Neo4jManager {
	driver, err := neo4j.NewDriver(fmt.Sprintf("bolt://%v:7687", neo4jHost), neo4j.BasicAuth("neo4j", "test", ""))
	if err != nil {
		panic("Neo4j driver failed to be instantiated")
	}
	return &Neo4jManager{driver}
}

func (m *Neo4jManager) Close() {
	m.driver.Close()
}

func (m *Neo4jManager) AddEdgeBetween(userID int64, nearbyUserID int64) error {
	_, err := m.executeQuery(fmt.Sprintf(`
			MATCH (a), (b)
			WHERE a.userid=%v AND b.userid=%v
			MERGE (a)-[r:%v]->(b)
		`, userID, nearbyUserID, relationshipIdentifier))
	if err != nil {
		return err
	}
	return nil
}

func (m *Neo4jManager) RetrieveUsersConnectedTo(userID int64) ([]int64, error) {
	records, err := m.executeQuery(fmt.Sprintf(`
			MATCH (a)
			WHERE a.userid=%v
			MATCH (a)-[r:%v*1..]-(b)
			RETURN DISTINCT b.userid as userid
		`, userID, relationshipIdentifier))
	if err != nil {
		return nil, err
	}

	return mapRecordsToUserIDs(records)
}

func mapRecordsToUserIDs(records []*neo4j.Record) ([]int64, error) {
	userIDs := make([]int64, len(records))
	for i, record := range records {
		userid, ok := record.Get("userid")
		if !ok {
			return nil, errors.New("expected key userid to be found in record")
		}
		userIDs[i] = userid.(int64)
	}
	return userIDs, nil
}

func (m *Neo4jManager) executeQuery(query string) ([]*neo4j.Record, error) {
	session := m.driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()
	result, err := session.Run(query, nil)
	if err != nil {
		return nil, err
	}

	records, err := result.Collect()
	if err != nil {
		return nil, err
	}
	return records, nil
}
