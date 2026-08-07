package data

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sync"
)

const (
	// neighborBBoxPad expands bboxes before overlap checks (degrees).
	neighborBBoxPad = 0.15
	// neighborEpsilon is the max distance between ring points to count as bordering.
	neighborEpsilon = 0.12
	// maxRingPoints caps samples per country for pairwise distance checks.
	maxRingPoints = 80
)

// Country is a slim metadata row for list endpoints.
type Country struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	ISO  string `json:"iso"`
}

// Store holds the loaded GeoJSON FeatureCollection and derived country list.
type Store struct {
	mu            sync.RWMutex
	rawJSON       json.RawMessage
	countries     []Country
	byID          map[int]Country
	neighborsByID map[int][]int
}

type featureCollection struct {
	Type     string    `json:"type"`
	Features []feature `json:"features"`
}

type feature struct {
	Type       string          `json:"type"`
	ID         int             `json:"id,omitempty"`
	Properties countryProps    `json:"properties"`
	Geometry   json.RawMessage `json:"geometry"`
}

type countryProps struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	ISO  string `json:"iso"`
}

type geoGeometry struct {
	Type        string          `json:"type"`
	Coordinates json.RawMessage `json:"coordinates"`
}

type countryShape struct {
	id     int
	minLng float64
	minLat float64
	maxLng float64
	maxLat float64
	points [][2]float64
}

// LoadCountries reads and indexes countries.geojson once at startup.
func LoadCountries(path string) (*Store, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read countries geojson: %w", err)
	}

	var fc featureCollection
	if err := json.Unmarshal(bytes, &fc); err != nil {
		return nil, fmt.Errorf("parse countries geojson: %w", err)
	}

	countries := make([]Country, 0, len(fc.Features))
	byID := make(map[int]Country, len(fc.Features))
	shapes := make([]countryShape, 0, len(fc.Features))

	for _, f := range fc.Features {
		c := Country{
			ID:   f.Properties.ID,
			Name: f.Properties.Name,
			ISO:  f.Properties.ISO,
		}
		countries = append(countries, c)
		byID[c.ID] = c

		shape, err := parseCountryShape(c.ID, f.Geometry)
		if err != nil {
			return nil, fmt.Errorf("parse geometry for country %d: %w", c.ID, err)
		}
		shapes = append(shapes, shape)
	}

	return &Store{
		rawJSON:       json.RawMessage(bytes),
		countries:     countries,
		byID:          byID,
		neighborsByID: buildNeighborIndex(shapes),
	}, nil
}

// GeoJSON returns the full FeatureCollection payload.
func (s *Store) GeoJSON() json.RawMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.rawJSON
}

// List returns country metadata.
func (s *Store) List() []Country {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Country, len(s.countries))
	copy(out, s.countries)
	return out
}

// HasCountry reports whether id exists in the store.
func (s *Store) HasCountry(id int) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.byID[id]
	return ok
}

// Neighbors returns land-adjacent countries for id (empty if none).
func (s *Store) Neighbors(id int) ([]Country, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.byID[id]; !ok {
		return nil, false
	}

	ids := s.neighborsByID[id]
	out := make([]Country, 0, len(ids))
	for _, nid := range ids {
		if c, ok := s.byID[nid]; ok {
			out = append(out, c)
		}
	}
	return out, true
}

func parseCountryShape(id int, raw json.RawMessage) (countryShape, error) {
	var geom geoGeometry
	if err := json.Unmarshal(raw, &geom); err != nil {
		return countryShape{}, err
	}

	shape := countryShape{
		id:     id,
		minLng: math.Inf(1),
		minLat: math.Inf(1),
		maxLng: math.Inf(-1),
		maxLat: math.Inf(-1),
	}

	var rings [][][]float64
	switch geom.Type {
	case "Polygon":
		var coords [][][]float64
		if err := json.Unmarshal(geom.Coordinates, &coords); err != nil {
			return countryShape{}, err
		}
		rings = coords
	case "MultiPolygon":
		var coords [][][][]float64
		if err := json.Unmarshal(geom.Coordinates, &coords); err != nil {
			return countryShape{}, err
		}
		for _, poly := range coords {
			rings = append(rings, poly...)
		}
	default:
		return countryShape{}, fmt.Errorf("unsupported geometry type %q", geom.Type)
	}

	var all [][2]float64
	for _, ring := range rings {
		for _, pt := range ring {
			if len(pt) < 2 {
				continue
			}
			lng, lat := pt[0], pt[1]
			if lng < shape.minLng {
				shape.minLng = lng
			}
			if lat < shape.minLat {
				shape.minLat = lat
			}
			if lng > shape.maxLng {
				shape.maxLng = lng
			}
			if lat > shape.maxLat {
				shape.maxLat = lat
			}
			all = append(all, [2]float64{lng, lat})
		}
	}

	shape.points = samplePoints(all, maxRingPoints)
	return shape, nil
}

func samplePoints(points [][2]float64, max int) [][2]float64 {
	if len(points) <= max {
		out := make([][2]float64, len(points))
		copy(out, points)
		return out
	}
	out := make([][2]float64, 0, max)
	step := float64(len(points)) / float64(max)
	for i := 0; i < max; i++ {
		idx := int(float64(i) * step)
		if idx >= len(points) {
			idx = len(points) - 1
		}
		out = append(out, points[idx])
	}
	return out
}

func buildNeighborIndex(shapes []countryShape) map[int][]int {
	index := make(map[int][]int, len(shapes))
	for i := 0; i < len(shapes); i++ {
		for j := i + 1; j < len(shapes); j++ {
			a, b := shapes[i], shapes[j]
			if !bboxOverlap(a, b, neighborBBoxPad) {
				continue
			}
			if !pointsNear(a.points, b.points, neighborEpsilon) {
				continue
			}
			index[a.id] = append(index[a.id], b.id)
			index[b.id] = append(index[b.id], a.id)
		}
	}
	return index
}

func bboxOverlap(a, b countryShape, pad float64) bool {
	return a.minLng-pad <= b.maxLng+pad &&
		a.maxLng+pad >= b.minLng-pad &&
		a.minLat-pad <= b.maxLat+pad &&
		a.maxLat+pad >= b.minLat-pad
}

func pointsNear(a, b [][2]float64, epsilon float64) bool {
	eps2 := epsilon * epsilon
	for _, pa := range a {
		for _, pb := range b {
			dLng := pa[0] - pb[0]
			dLat := pa[1] - pb[1]
			if dLng*dLng+dLat*dLat <= eps2 {
				return true
			}
		}
	}
	return false
}
