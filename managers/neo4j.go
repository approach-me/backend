package managers

import (
	"errors"
	"fmt"
	"os"

	"github.com/approach.me/backend/protos"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

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

func (m *Neo4jManager) CreateOrUpdateNode(userSummary *protos.UserSummary) error {
	_, err := m.executeQuery(fmt.Sprintf(`
			MERGE ({
				userid: "%v",
				deviceid: "%v",
				name: "%v",
				birthdate: %v,
				thumbnail: "%v"
			})
		`, userSummary.UserId, userSummary.DeviceId, userSummary.Name, userSummary.Birthdate, userSummary.ThumbnailUri))
	if err != nil {
		return err
	}
	return nil
}

func (m *Neo4jManager) AddEdgeBetween(userID, nearbyUserID string) error {
	_, err := m.executeQuery(fmt.Sprintf(`
			MATCH (a), (b)
			WHERE a.userid="%v" AND b.deviceid="%v"
			MERGE (a)-[r:CONNECTED]->(b)
		`, userID, nearbyUserID))
	if err != nil {
		return err
	}
	return nil
}

func (m *Neo4jManager) UnlinkEdgesAndMergeDanglingComponentsFrom(userID string) error {
	_, err := m.executeQuery(fmt.Sprintf(`
		MATCH (n {userid:"%v"})
		OPTIONAL MATCH (c1)-[r1:CONNECTED]-(n)-[r2:CONNECTED]-(c2)
		DELETE r1, r2
		WITH c1, c2 WHERE c2 IS NOT NULL AND c1 IS NOT NULL
		MERGE (c1)-[:CONNECTED]-(c2)
		`, userID))
	if err != nil {
		return err
	}
	return nil
}

func (m *Neo4jManager) RetrieveUsersConnectedTo(userID string) ([]*protos.UserSummary, error) {
	records, err := m.executeQuery(fmt.Sprintf(`
			MATCH p = ({userid: "%v"})-[*]-()
			UNWIND NODES(p) as userNode
			RETURN DISTINCT userNode{.*}
			ORDER BY userNode.userid
		`, userID))
	if err != nil {
		return nil, err
	}

	return mapRecordsToUserSummaries(records)
}

func mapRecordsToUserSummaries(records []*neo4j.Record) ([]*protos.UserSummary, error) {
	userSummaries := make([]*protos.UserSummary, len(records))
	for i, record := range records {
		v, ok := record.Get("userNode")
		if !ok {
			return nil, errors.New("expected key userNode to be found in record")
		}
		userNode := v.(map[string]interface{})
		userSummaries[i] = &protos.UserSummary{
			UserId:       userNode["userid"].(string),
			DeviceId:     userNode["deviceid"].(string),
			Name:         userNode["name"].(string),
			Birthdate:    userNode["birthdate"].(int64),
			ThumbnailUri: userNode["thumbnail"].(string),
		}
	}
	return userSummaries, nil
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
