/*
 *  Remote Desktop Services
 *
 * Documentation automatically generated by the <b>swagger-autogen</b> module.
 *
 * API version: 1.0.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package swagger

type UpdateStatus struct {
	State string `json:"state,omitempty"`
	Rx float32 `json:"rx,omitempty"`
	Tx float32 `json:"tx,omitempty"`
}