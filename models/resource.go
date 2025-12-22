package models

// ResourceDefinition represents a FHIR resource definition with its elements
type ResourceDefinition struct {
	ResourceType string      `json:"resourceType,omitempty"`
	Name         string      `json:"name"`
	Flags        []string    `json:"flags,omitempty"`
	Type         string      `json:"type"`
	Description  string      `json:"description,omitempty"`
	Elements     []Element   `json:"elements,omitempty"`
	Extensions   []Extension `json:"extensions,omitempty"`
}

// Element represents a single element/field in the resource definition
type Element struct {
	Name        string      `json:"name"`
	Flags       []string    `json:"flags,omitempty"`
	Cardinality string      `json:"cardinality,omitempty"`
	Type        string      `json:"type"`
	TypeRef     string      `json:"typeRef,omitempty"`
	Description string      `json:"description,omitempty"`
	Usage       string      `json:"usage,omitempty"`       // "used", "not-used", "todo", "optional"
	Notes       string      `json:"notes,omitempty"`       // Custom implementation notes
	Binding     *Binding    `json:"binding,omitempty"`     // Value set binding
	Elements    []Element   `json:"elements,omitempty"`    // Nested child elements
	Extensions  []Extension `json:"extensions,omitempty"`  // Extensions on this element
}

// Binding represents a value set binding for coded elements
type Binding struct {
	Strength string `json:"strength,omitempty"` // "required", "extensible", "preferred", "example"
	ValueSet string `json:"valueSet,omitempty"` // Value set URL or pipe-delimited values
	URL      string `json:"url,omitempty"`      // Link to value set documentation
}

// Extension represents a FHIR extension definition
type Extension struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Context     string `json:"context,omitempty"`     // Where extension can be used
	Type        string `json:"type"`
	Cardinality string `json:"cardinality,omitempty"` // Cardinality like "0..1"
	Description string `json:"description,omitempty"`
}

// Flag constants for FHIR element flags
const (
	FlagSummary    = "S"   // Σ - Summary element
	FlagModifier   = "?!"  // Modifier element
	FlagConstraint = "I"   // Has constraint
	FlagTrialUse   = "TU"  // Trial use
	FlagNormative  = "N"   // Normative
)

// Usage constants
const (
	UsageUsed    = "used"
	UsageNotUsed = "not-used"
	UsageTodo    = "todo"
	UsageOptional = "optional"
)

// FlatElement represents a flattened element with depth info for rendering
type FlatElement struct {
	Element     Element
	Depth       int
	IsLast      bool     // Is this the last child at its depth
	ParentLasts []bool   // Track if ancestors were last children (for tree lines)
	Path        string   // Full path like "participant.type"
}

// Flatten recursively flattens the element hierarchy for rendering
func (r *ResourceDefinition) Flatten() []FlatElement {
	var result []FlatElement

	// Add root element
	rootElement := Element{
		Name:        r.Name,
		Flags:       r.Flags,
		Type:        r.Type,
		Description: r.Description,
	}
	result = append(result, FlatElement{
		Element:     rootElement,
		Depth:       0,
		IsLast:      len(r.Elements) == 0 && len(r.Extensions) == 0,
		ParentLasts: []bool{},
		Path:        r.Name,
	})

	// Flatten children
	flattenElements(r.Elements, 1, &result, []bool{}, r.Name, false)

	// Add extensions at the end
	for i, ext := range r.Extensions {
		extElement := Element{
			Name:        ext.Name,
			Type:        ext.Type,
			Description: ext.Description,
		}
		isLast := i == len(r.Extensions)-1
		result = append(result, FlatElement{
			Element:     extElement,
			Depth:       1,
			IsLast:      isLast,
			ParentLasts: []bool{len(r.Elements) == 0},
			Path:        ext.Context,
		})
	}

	return result
}

func flattenElements(elements []Element, depth int, result *[]FlatElement, parentLasts []bool, parentPath string, parentIsLast bool) {
	for i, elem := range elements {
		isLast := i == len(elements)-1
		path := parentPath + "." + elem.Name

		newParentLasts := make([]bool, len(parentLasts)+1)
		copy(newParentLasts, parentLasts)
		newParentLasts[len(parentLasts)] = parentIsLast

		*result = append(*result, FlatElement{
			Element:     elem,
			Depth:       depth,
			IsLast:      isLast && len(elem.Extensions) == 0,
			ParentLasts: newParentLasts[:len(newParentLasts)-1],
			Path:        path,
		})

		if len(elem.Elements) > 0 {
			flattenElements(elem.Elements, depth+1, result, newParentLasts, path, isLast && len(elem.Extensions) == 0)
		}

		// Add extensions nested under this element
		for j, ext := range elem.Extensions {
			extElement := Element{
				Name:        ext.Name,
				Type:        ext.Type,
				Cardinality: ext.Cardinality,
				Description: ext.Description,
			}
			extIsLast := j == len(elem.Extensions)-1 && isLast
			*result = append(*result, FlatElement{
				Element:     extElement,
				Depth:       depth + 1,
				IsLast:      extIsLast,
				ParentLasts: newParentLasts,
				Path:        path + "." + ext.Name,
			})
		}
	}
}
