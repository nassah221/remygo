/*
 *  Remote Desktop Services
 *
 * Documentation automatically generated by the <b>swagger-autogen</b> module.
 *
 * API version: 1.0.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package swagger

type Session struct {
	Id string `json:"id,omitempty"`
	Identifier string `json:"identifier,omitempty"`
	State string `json:"state,omitempty"`
	EndTime string `json:"endTime,omitempty"`
	Rx float32 `json:"rx,omitempty"`
	Tx float32 `json:"tx,omitempty"`
	StartTime string `json:"startTime,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
	DeviceId string `json:"DeviceId,omitempty"`
	Devices []SessionDevices `json:"devices,omitempty"`
	SessionLogs []SessionLog `json:"SessionLogs,omitempty"`
}
