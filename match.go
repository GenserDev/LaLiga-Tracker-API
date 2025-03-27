package main

import (
	"database/sql"
	"time"
)

type Match struct {
	ID             int       `json:"id"`
	HomeTeam       string    `json:"homeTeam"`
	AwayTeam       string    `json:"awayTeam"`
	MatchDate      time.Time `json:"matchDate"`
	Goals          int       `json:"goals"`
	YellowCards    int       `json:"yellowCards"`
	RedCards       int       `json:"redCards"`
	ExtraTime      bool      `json:"extraTime"`
}

func (m *Match) Create(db *sql.DB) error {
	query := `INSERT INTO matches 
		(home_team, away_team, match_date, goals, yellow_cards, red_cards, extra_time) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	
	return db.QueryRow(
		query, 
		m.HomeTeam, 
		m.AwayTeam, 
		m.MatchDate, 
		m.Goals, 
		m.YellowCards, 
		m.RedCards, 
		m.ExtraTime,
	).Scan(&m.ID)
}

func GetMatchByID(db *sql.DB, id int) (*Match, error) {
	m := &Match{}
	query := `SELECT id, home_team, away_team, match_date, 
			  goals, yellow_cards, red_cards, extra_time 
			  FROM matches WHERE id = $1`
	
	err := db.QueryRow(query, id).Scan(
		&m.ID, 
		&m.HomeTeam, 
		&m.AwayTeam, 
		&m.MatchDate, 
		&m.Goals, 
		&m.YellowCards, 
		&m.RedCards, 
		&m.ExtraTime,
	)
	
	if err != nil {
		return nil, err
	}
	return m, nil
}

func GetAllMatches(db *sql.DB) ([]Match, error) {
	query := `SELECT id, home_team, away_team, match_date, 
			  goals, yellow_cards, red_cards, extra_time 
			  FROM matches`
	
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []Match
	for rows.Next() {
		var m Match
		err := rows.Scan(
			&m.ID, 
			&m.HomeTeam, 
			&m.AwayTeam, 
			&m.MatchDate, 
			&m.Goals, 
			&m.YellowCards, 
			&m.RedCards, 
			&m.ExtraTime,
		)
		if err != nil {
			return nil, err
		}
		matches = append(matches, m)
	}
	return matches, nil
}

func (m *Match) Update(db *sql.DB) error {
	query := `UPDATE matches SET 
		home_team = $1, 
		away_team = $2, 
		match_date = $3,
		goals = $4,
		yellow_cards = $5,
		red_cards = $6,
		extra_time = $7
		WHERE id = $8`
	
	_, err := db.Exec(
		query, 
		m.HomeTeam, 
		m.AwayTeam, 
		m.MatchDate, 
		m.Goals, 
		m.YellowCards, 
		m.RedCards, 
		m.ExtraTime,
		m.ID,
	)
	return err
}

func DeleteMatch(db *sql.DB, id int) error {
	query := `DELETE FROM matches WHERE id = $1`
	_, err := db.Exec(query, id)
	return err
}