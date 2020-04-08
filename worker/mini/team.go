package main

import (
    "log"
)

type Team struct{
    teamID      int
    selfType    int
    hasRobot    bool
    allPlayers  map[int]*Player
}

func (this *Team)AddPlayer(player *Player){
    this.allPlayers[player.userID] = player
}
func (this *Team)GetPlayer(userID int) *Player{
    return this.allPlayers[userID]
}
func (this *Team)DelPlayer(userID int) bool{
    if this.allPlayers[userID] != nil {
        delete(this.allPlayers, userID)
        return true
    }
    return false
}
func (this *Team)GetMemberNum() int{
    return len(this.allPlayers)
}
func (this *Team)Foreach(fn func(player *Player)){
    log.Println(this.allPlayers)
    for _, player := range this.allPlayers{
        fn(player)
    }
}

var gAllTeams = make(map[int]*Team)

func NewTeam(id int) *Team{
    return &Team{id, 0, false, make(map[int]*Player)}
}

func AddTeam(team *Team){
    gAllTeams[team.teamID] = team
}

func GetTeam(id int) *Team{
    return gAllTeams[id]
}

func DelTeam(teamID int) bool{
    if gAllTeams[teamID] != nil {
        delete(gAllTeams, teamID)
        return true
    }
    return false
}

func ForeachTeam(fn func(team *Team)){
    for _, tea := range gAllTeams{
        fn(tea)
    }
}
