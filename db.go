package apron

import (
	"github.com/evq/go-zigbee"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	gorm.DB
}

type masterDevice struct {
	ID       int64    `gorm:"column:deviceId;primary_key"`
	UserName string `gorm:"column:userName;type:varchar(100)"`
}

func (m masterDevice) TableName() string { return "masterDevice" }

// FIXME nulls?
type aZigbeeDevice struct {
	ID               int64 `gorm:"column:masterId"`
	IeeeAddr         int64 `gorm:"column:globalId"`
	NetAddr          int64 `gorm:"column:networkId"`
	EndpointId       int64 `gorm:"column:endpointId"`
	DeviceType       int64 `gorm:"column:deviceType"`
	ManufacturerCode int64 `gorm:"column:manufacturerCode"`
	ProductId        int64 `gorm:"column:productId"`
}

func (z aZigbeeDevice) TableName() string { return "zigbeeDevice" }

type zigbeeDeviceState struct {
	IeeeAddr            int64    `gorm:"column:globalId;primary_key"`
	EndpointId          int64    `gorm:"column:endpointId"`
	ClusterId           int64    `gorm:"column:clusterId"`
	Attribute           int64    `gorm:"column:attributeId"`
	ValueGet            string `gorm:"type:varchar(256)"`
	ValueSet            string `gorm:"type:varchar(256)"`
	SetValueChangedFlag bool   `gorm:"column:setValueChangedFlag"`
	LastUpdate          int64    `gorm:"column:lastUpdate"`
}

func (z zigbeeDeviceState) TableName() string { return "zigbeeDeviceState" }

func Open(path string) (*DB, error) {
	if path == "" {
		path = "./apron.db"
	}
	db, err := gorm.Open("sqlite3", path)
	return &DB{db}, err
}

func (db *DB) GetZigbeeDevices() []zigbee.ZigbeeDevice {
	z := []aZigbeeDevice{}
	db.Find(&z)

	devs := make([]zigbee.ZigbeeDevice, len(z), len(z))
	for i := range z {
		devs[i].NetAddr = uint16(z[i].NetAddr)
		devs[i].IeeeAddr = uint64(z[i].IeeeAddr)
		devs[i].ManufacturerCode = uint16(z[i].ManufacturerCode)
		endptid := uint8(z[i].EndpointId)
		devs[i].Endpoints = make(map[uint8]*zigbee.Endpoint)
		devs[i].Endpoints[endptid] = &zigbee.Endpoint{}
		devs[i].Endpoints[endptid].ID = endptid
		devs[i].Endpoints[endptid].DeviceType = uint16(z[i].DeviceType)

		m := masterDevice{}
		db.Model(&z[i]).Related(&m, "ID")
		devs[i].Name = m.UserName

		st := []zigbeeDeviceState{}
		db.Model(&z[i]).Related(&st, "IeeeAddr")
		devs[i].Endpoints[endptid].InClusters = make(map[uint16]*zigbee.Cluster)
		for j := range st {
			clusterid := uint16(st[j].ClusterId)
			if devs[i].Endpoints[endptid].InClusters[clusterid] == nil {
				devs[i].Endpoints[endptid].InClusters[clusterid] = &zigbee.Cluster{}
				devs[i].Endpoints[endptid].InClusters[clusterid].ID = clusterid
				devs[i].Endpoints[endptid].InClusters[clusterid].Attributes = make(map[uint16]string)
			}
			attrid := uint16(st[j].Attribute & 0x0FFF)
			devs[i].Endpoints[endptid].InClusters[clusterid].Attributes[attrid] = st[j].ValueGet
		}
	}
	return devs
}
