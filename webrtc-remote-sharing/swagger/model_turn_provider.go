/*
 *  Remote Desktop Services
 *
 * Documentation automatically generated by the <b>swagger-autogen</b> module.
 *
 * API version: 1.0.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package swagger

type TurnProvider struct {
	Id string `json:"id,omitempty"`
	Description string `json:"description"`
	Url string `json:"url"`
	Port float32 `json:"port"`
	Protocol string `json:"protocol"`
	User string `json:"user"`
	Password string `json:"password"`
	UpdatedAt string `json:"updatedAt,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
}
