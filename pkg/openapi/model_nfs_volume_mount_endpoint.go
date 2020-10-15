/*
 * StorageOS API
 *
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * API version: 2.3.0-alpha
 * Contact: info@storageos.com
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// NfsVolumeMountEndpoint struct for NfsVolumeMountEndpoint
type NfsVolumeMountEndpoint struct {
	// The address to which the NFS server is bound.
	MountEndpoint string `json:"mountEndpoint,omitempty"`
	// An opaque representation of an entity version at the time it was obtained from the API. All operations that mutate the entity must include this version field in the request unchanged. The format of this type is undefined and may change but the defined properties will not change.
	Version string `json:"version,omitempty"`
}
