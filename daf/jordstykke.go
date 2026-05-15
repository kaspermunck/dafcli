package daf

import "fmt"

// Jordstykke is a single MAT_Jordstykke node (a cadastral parcel).
//
// Field naming follows Datafordeler GraphQL schema verbatim — note the
// capital L in id_lokalId. samletFastEjendomLokalId is the FK to the parcel's
// SFE; the SFE's id_lokalId is the BFE-nummer.
type Jordstykke struct {
	IDLokalId                string `json:"id_lokalId"`
	Matrikelnummer           string `json:"matrikelnummer"`
	RegistreretAreal         int    `json:"registreretAreal"`
	Status                   string `json:"status"`
	SamletFastEjendomLokalId string `json:"samletFastEjendomLokalId"`
	EjerlavLokalId           string `json:"ejerlavLokalId"`
	RegistreringFra          string `json:"registreringFra,omitempty"`
	VirkningFra              string `json:"virkningFra,omitempty"`
}

// BFE returns the parcel's BFE-nummer (= the SFE's id_lokalId).
// Empty if the SFE relation isn't set on this Jordstykke.
func (j Jordstykke) BFE() string { return j.SamletFastEjendomLokalId }

type jordstykkeData struct {
	MatJordstykke struct {
		Nodes []Jordstykke `json:"nodes"`
	} `json:"MAT_Jordstykke"`
}

// Jordstykker looks up parcels by matrikelnummer, optionally narrowed by
// ejerlavLokalId (use the integer ejerlavskode from DAWA as a string).
// Returns up to `limit` matches at the current point in time.
func (c *Client) Jordstykker(matrikelnummer, ejerlavLokalId string, limit int) ([]Jordstykke, error) {
	if limit <= 0 {
		limit = 10
	}
	ts := NowTimestamp()

	where := fmt.Sprintf(`{matrikelnummer:{eq:%q}}`, matrikelnummer)
	if ejerlavLokalId != "" {
		where = fmt.Sprintf(`{matrikelnummer:{eq:%q}, ejerlavLokalId:{eq:%q}}`, matrikelnummer, ejerlavLokalId)
	}

	query := fmt.Sprintf(`{
		MAT_Jordstykke(registreringstid:%q, virkningstid:%q, where:%s, first:%d) {
			nodes {
				id_lokalId
				matrikelnummer
				registreretAreal
				status
				samletFastEjendomLokalId
				ejerlavLokalId
				registreringFra
				virkningFra
			}
		}
	}`, ts, ts, where, limit)

	raw, err := c.QueryRaw("MAT", query)
	if err != nil {
		return nil, err
	}
	data, err := decodeGraphQL[jordstykkeData](raw)
	if err != nil {
		return nil, err
	}
	return data.MatJordstykke.Nodes, nil
}

// SFE is a Samlet Fast Ejendom node. Its id_lokalId IS the BFE-nummer.
type SFE struct {
	IDLokalId        string `json:"id_lokalId"`
	Status           string `json:"status"`
	DatafordelerRowId string `json:"datafordelerRowId,omitempty"`
}

// BFE returns the SFE's id_lokalId (its BFE-nummer).
func (s SFE) BFE() string { return s.IDLokalId }

type sfeData struct {
	MatSFE struct {
		Nodes []SFE `json:"nodes"`
	} `json:"MAT_SamletFastEjendom"`
}

// SFEByBFE looks up a Samlet Fast Ejendom by its BFE-nummer (= id_lokalId).
func (c *Client) SFEByBFE(bfeNummer string) (*SFE, error) {
	ts := NowTimestamp()
	query := fmt.Sprintf(`{
		MAT_SamletFastEjendom(registreringstid:%q, virkningstid:%q, where:{id_lokalId:{eq:%q}}, first:1) {
			nodes { id_lokalId status datafordelerRowId }
		}
	}`, ts, ts, bfeNummer)

	raw, err := c.QueryRaw("MAT", query)
	if err != nil {
		return nil, err
	}
	data, err := decodeGraphQL[sfeData](raw)
	if err != nil {
		return nil, err
	}
	if len(data.MatSFE.Nodes) == 0 {
		return nil, fmt.Errorf("no SFE found for BFE %s", bfeNummer)
	}
	return &data.MatSFE.Nodes[0], nil
}
