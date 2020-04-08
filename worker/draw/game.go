package main

import (

)

type Game struct {
    gameID   int   
    mapObj   *MapObj   
    state   string   
    painterIndex   int   
    wordIndex      int
    painterCount   int   
    selectTime   int   
    drawTime   int   
    showTime   int   
    endTime   int   
    playNum   int   
    overTime   int   
    words   []map[string]interface{}
    indics   []int
    commands   []map[string]interface{}
    answers   []map[string]interface{}
    scores   []int
}

var gAllGames = make(map[int]*Game)

func AddGame(gameID int, game *Game) {
    gAllGames[gameID] = game   
}
func GetGame(gameID int) *Game{
    return gAllGames[gameID]   
}
func DelGame(gameID int) bool{
    if gAllGames[gameID] != nil{
        delete(gAllGames, gameID)   
        return true   
    }
    return false   
}
func ForeachGame(fn func(game *Game)) {
    for _, game := range gAllGames {
        fn(game)
    }
    return
}
func NewGame(id int, n int) *Game{
    return &Game{id,nil,"idle",-1,-1,0,0,0,0,0,0,0,
        make([]map[string]interface{}, 0),
        make([]int,0),
        make([]map[string]interface{}, 0),
        make([]map[string]interface{}, 0),
        make([]int,0),
        }  
}