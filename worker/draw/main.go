package main

import (
    "bufio"
    "log"
    "os"
    "math"
    "math/rand"
    "io/ioutil"
    "time"
    "encoding/json"

    "easyserver/util"
    "easyserver/util/db"
    "easyserver/lib/worker"
)

func getLevel(exp int) int{
    x := exp / 40
    if x < 1 {
        return 1
    }
    return int(math.Log2(float64(x))) + 2
}

var vocabularies []map[string]interface{}
type JsonStruct struct {}

func loadJson() {
    jsonParse := &JsonStruct{}
    jsonParse.Load("./vocabularies.json", &vocabularies)
}

func (jst *JsonStruct) Load(filename string, v interface{}) {
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return
    }
    err = json.Unmarshal(data, v)
    if err != nil {
        return
    }
}

func getMembers(team *Team) map[string]interface{}{
    players := make([]map[string]interface{}, 0)
    obsers := make([]map[string]interface{}, 0)

    team.Foreach(func(playerInTeam *Player){
        info := map[string]interface{}{
            "id": playerInTeam.userID,
            "user": playerInTeam.userName,
            "nick": playerInTeam.userName,
            "online": playerInTeam.online,
            "ready": playerInTeam.ready,
            "gender": playerInTeam.gender,
            "avatar": playerInTeam.avatar,
            "score": playerInTeam.score,
            "exp": playerInTeam.exp,
            "isplayer": playerInTeam.isPlayer,
        }
        if playerInTeam.isPlayer {
            players = append(players, info)
        }else{
            obsers = append(obsers, info)

        }
    })

    return map[string]interface{}{ "players": players, "obsers": obsers }
}

func enterGame(player *Player, team *Team) {
    var info map[string]interface{}
    game := GetGame(team.teamID)
    if game == nil{
        player.isPlayer = team.GetPlayerNum() <= MAX_NUM_EACH_TEAM
    }else if game.state == "selecting"{
        info = map[string]interface{}{
            "state": "selecting",
            "idx": game.painterIndex,
            "t": game.selectTime,
            "words": game.words,
        }
    }else if game.state == "drawing"{
        info = map[string]interface{}{
            "state": "drawing",
            "idx": game.painterIndex,
            "t": game.drawTime,
            "word": game.words[game.wordIndex]["word"],
            "tip": game.words[game.wordIndex]["tip"],
            "cmds": game.commands,
        }
    }else if game.state == "showing"{
        info = map[string]interface{}{
            "state": "showing",
            "idx": game.painterIndex,
            "t": game.showTime,
            "word": game.words[game.wordIndex]["word"],
            "cmds": game.commands,
        }
    }else if game.state == "over"{
        info = map[string]interface{}{
            "state": "over",
            "t": game.overTime,
            "scores": []int{},
        }
    }
        

    members := getMembers(team)
    data := map[string]interface{}{
        "s": 1,
        "id": player.team.teamID,
        "crter": player.team.creator,
        "isplayer": player.isPlayer,
        "members": members,
        "info": info,
    }
    player.SendMsg("game.enter", data)
    player.NotifyOtherInTeam("user.new", map[string]interface{}{
        "members": members, "id": player.userID, "isplayer": player.isPlayer,
    })

}
func doGameOver(gameID int) {
    game := GetGame(gameID)
    if game == nil{
        return
    }

    team := GetTeam(gameID)
    team.Broadcast("game.over", map[string]interface{}{ "rsts": game.scores })


    team.ForeachPlayer(func (index int, player *Player) {
        if (player.online == false) {
            DelPlayer(player.GetID())
            player.team.DelPlayer(player.userID)
            mapObj := player.mapObj
            mapObj.DelPlayer(player.userID)
            player.mapObj = nil
        }else{
            player.ready = false
        }
        player.score = game.scores[player.team.GetPlayerIndex(player.userID)]
        add := util.If(player.score < 0 , 0, player.score).(int)
        player.exp += add
        player.level = getLevel(player.exp)
        db.UpdateScore(player.userID, GAME_ID, player.score)
        db.AddExp(player.userID, GAME_ID, add)
    })
    DelGame(gameID)
    game.mapObj.DelGame(gameID)
    game.mapObj = nil
}

func gen(game *Game, words []map[string]interface{}, indics []int) []map[string]interface{}{
    index := int(rand.Intn(1000) / 1000.0 * len(vocabularies))
    for i := 0; i < len(indics); i+=1{
        if index == indics[i]{
            return gen(game, words, indics)
        }
    }
    for i := 0; i < len(game.indics); i+=1{
        if index == indics[i]{
            return gen(game, words, indics)
        }
    }
    v := vocabularies[index]
    indics = append(indics,index)
    words = append(words,map[string]interface{}{ "tip": v["type"], "word": v["word"], "idx": index, })

    if len(words) < 4{
        return gen(game, words, indics)
    }else{
        return words
    }
}

func genWords(gameID int) []map[string]interface{}{
    game := GetGame(gameID)
    if game == nil{
        return nil
    }

    words := make([]map[string]interface{}, 0)
    indics := make([]int, 0)
    
    return gen(game, words, indics)
}

func selectWord(gameID int, quick bool) {
    game := GetGame(gameID)
    if game == nil{
        return
    }

    team := GetTeam(gameID)

    if game.painterCount < game.playNum {
        game.state = "selecting"
        game.selectTime = int(time.Now().Unix())
        game.painterIndex = team.getPainterIndex(game.painterCount)
        game.painterCount += 1
        if game.painterIndex == -1{
            return
        }
    }else {
        doGameOver(gameID)
        return
    }

    util.SetTimeout(func () {
        game.selectTime = 0
        if game.wordIndex == -1{
            selectWord(gameID, true)
        }else {
            game.drawTime = int(time.Now().Unix())
            data := map[string]interface{}{
                "idx": game.painterIndex,
                "time": game.drawTime,
                "word": game.words[game.wordIndex]["word"],
                "tip": game.words[game.wordIndex]["tip"],
            }
            game.state = "drawing"
            game.indics = append(game.indics,game.words[game.wordIndex]["idx"].(int))
            team.Broadcast("game.draw", data)

            util.SetTimeout(func () {
                showAnswer(gameID)
            },60000)
        }
    },5000)

    game.words = genWords(gameID)
    game.commands = make([]map[string]interface{},0)
    data := map[string]interface{}{
        "idx": game.painterIndex,
        "time": game.selectTime,
        "words": game.words,
        "quick": quick,
    }
    team.Broadcast("game.select", data)
}

func showAnswer(gameID int) {
    game := GetGame(gameID)
    if game == nil{
        return
    }

    team := GetTeam(gameID)
    game.state = "counting"
    game.showTime = int(time.Now().Unix())

    point := 0
    num := 0
    for i := 0; i < len(game.answers); i+=1{
        if game.answers[i]["answer"] == game.words[game.wordIndex]["word"].(string) {
            point += 2
            num += 2
        }
    }
    if num >= 5{
        point = 0
    }

    data := map[string]interface{}{
        "num": num,
        "answer": game.words[game.wordIndex]["word"],
        "point": point,
        "score": game.scores[game.painterIndex],
    }

    team.Broadcast("game.result", data)

    util.SetTimeout(func () {
        game.state = "showing"
        team.Broadcast("game.answer", nil)
    },3000)

    util.SetTimeout(func () {
        game.showTime = 0
        game.answers = make([]map[string]interface{}, 0)
        game.wordIndex = -1
        selectWord(gameID, false)
    },8000)
}

func commitAnswer(player *Player, answer string) {
    game := GetGame(player.team.teamID)
    if game == nil && game.state != "drawing"{
        return
    }

    team := player.team
    index := team.GetPlayerIndex(player.userID)
    if index == game.painterIndex{
        return
    }
    for i := 0; i < len(game.answers); i+=1{
        if answer == game.answers[i]["answer"].(string) && index == i{
            return
        }
    }
    scores := []int{7, 5, 4, 3, 2}
    score := util.If(answer != game.words[game.wordIndex]["word"], 0, scores[len(game.answers)]).(int)
    game.answers = append(game.answers,map[string]interface{}{ "userId": player.userID, "answer": answer })
    game.scores[index] += score
    game.scores[game.painterIndex] += 2

    if len(game.answers) >= 3 {
        num := 0
        for i := 0; i < len(game.answers); i+=1{
            if game.answers[i]["answer"] == game.words[game.wordIndex]["word"]{
                num += 1
            }
        }
        if num == 5{
            game.scores[game.painterIndex] -= 21
        }
    }
    data := map[string]interface{}{
        "id": player.userID,
        "answer": answer,
        "score": score,
    }
    player.BroadcastInTeam("game.commit", data)
}

var gMsgCallBack = make(map[string]func(player *Player, data map[string]interface{}))
func bind(name string, fn func(player *Player, data map[string]interface{})){
    gMsgCallBack[name] = fn
}

var gIndexPlayerID = 1

func loginCb(player *Player, userID int, data map[string]interface{}){
    if userID < 0 {
        return
    }
    player.userID = userID
    gIndexPlayerID += 1
    AddPlayer(player)
    player.mapObj = AllocMap()
    player.mapObj.AddPlayer(player)
    player.playerID = player.mapObj.AllocPlayerID()
    db.GetScore(userID, GAME_ID, func(score int){
        player.score = score
        player.score = util.If(player.score < 0 , 0 , player.score).(int)
        db.GetExp(userID, GAME_ID, func(exp int){
            player.exp = exp
            player.exp = util.If(player.exp < 0 , 0 , player.exp).(int)
            player.level = getLevel(player.exp)
            token := util.NewToken(userID, 7*24*3600*1000)
            log.Println("send init data")
            player.SendMsg("game.init", map[string]interface{}{"token":token,"t": time.Now().Unix(), "id": userID,
                                    "score": player.score, "exp": player.exp, "lv": player.level})
        })
    })
}

func bindHandler(){
    bind("login", func(player *Player, data map[string]interface{}){
        id, ok := data["id"].(float64)
        if !ok {
            id = -1
        }
        userID := int(id)
        player.userName = data["user"].(string)
        player.nickName = data["nick"].(string)
        player.avatar = data["avatar"].(string)
        player.gender = int(data["gender"].(float64))

        if userID == -1{
            db.CreateUser(player.userName,player.nickName,player.gender,player.avatar, func(userid int){loginCb(player, userid, data)})
        }
        if userID == -1{
            return
        }
        go func(userid int){loginCb(player, userid, data)}(gIndexPlayerID)
    })

    bind("invite", func(player *Player, data map[string]interface{}) {
        team := player.team
        create := false
        if team == nil {
            team = player.mapObj.AllocTeam(player)
            team.creator = player.userID
            create = true
        }

        player.BroadcastInMap(
            "player.invite", map[string]interface{}{ "n": player.nickName, "id": player.userID, "team": team.teamID })

        if create{
            enterGame(player, team)
        }
    })

    bind("enter", func(player *Player, data map[string]interface{}) {
        var team = GetTeam(int(data["id"].(float64)))
        var game = GetGame(team.teamID)
        if player.team != nil || team == nil || team.GetMemberNum() >= MAX_NUM_EACH_TEAM + MAX_OB_EACH_TEAM{
            return
        }else if game != nil && team.GetMemberNum() < MAX_NUM_EACH_TEAM {
            player.team = team
            team.AddObser(player)
        }else {
            player.team = team
            team.AddPlayer(player)
        }
        enterGame(player, team)
    })

    bind("ready", func(player *Player, data map[string]interface{}) {
        if player.ready{
            return
        }
        player.ready = true
        player.BroadcastInTeam("player.ready", map[string]interface{}{ "id": player.userID })
        n := player.team.GetPlayerNum()
        if n > 1 && player.team.GetReadyNum() == n{
            player.BroadcastInTeam("game.allready", nil)
        }
        util.SetTimeout(func () {
            game := NewGame(player.team.teamID, player.team.GetPlayerNum())
            game.mapObj = player.mapObj
            AddGame(game.gameID, game)
            player.mapObj.AddGame(game.gameID, game)
            game.painterIndex = 0
            player.BroadcastInTeam("game.begin", map[string]interface{}{ "idx": game.painterIndex })
            game.state = "begin"
            game.playNum = n
            selectWord(game.gameID, false)
        },5000)
    })

    bind("standup", func(player *Player, data map[string]interface{}) {
        rst := player.team.Standup(player)
        if rst {
            members := getMembers(player.team)
            player.BroadcastInTeam(
                "user.standup", map[string]interface{}{ "id": player.userID, "members": members })
        }
    })

    bind("sitdown", func(player *Player, data map[string]interface{}) {
        rst := player.team.Standup(player)
        if rst {
            members := getMembers(player.team)
            player.BroadcastInTeam(
                "user.sitdown", map[string]interface{}{ "id": player.userID, "members": members })
        }
    })

    bind("tool", func(player *Player, data map[string]interface{}) {
        game := GetGame(player.team.teamID)
        if game == nil{
            return
        }
        game.commands = append(game.commands,data)
        player.NotifyOtherInTeam("cmd.tool", data)
    })

    bind("shape", func(player *Player, data map[string]interface{}) {
        game := GetGame(player.team.teamID)
        if game == nil{
            return
        }
        game.commands = append(game.commands,map[string]interface{}{ "type": "draw", "data": data })
        player.NotifyOtherInTeam("cmd.shape", data)
    })

    bind("select", func(player *Player, data map[string]interface{}) {
        game := GetGame(player.team.teamID)
        if game == nil{
            return
        }
        game.wordIndex = data["idx"].(int)
    })

    bind("chat", func(player *Player, data map[string]interface{}) {
        player.NotifyOtherInTeam("game.chat", data)
    })


    bind("commit", func(player *Player, data map[string]interface{}) {
        commitAnswer(player, data["str"].(string))
    })


    bind("cancel", func(player *Player, data map[string]interface{}) {
        player.ready = true
        player.BroadcastInTeam("player.cancel", map[string]interface{}{ "id": player.userID })
    })

    bind("kick", func(player *Player, data map[string]interface{}) {
        target := GetPlayer(data["id"].(int))
        if target != nil {
            player.team.DelPlayer(data["id"].(int))
            target.team = nil
            player.team.Broadcast("user.state", map[string]interface{}{ "id": data["id"], "online": false, "members": getMembers(player.team) })
            target.SendMsg("game.kick", nil)
        }
    })

    bind("exit", func(player *Player, data map[string]interface{}) {
        team := player.team
        userID := player.userID
        if team.creator == userID{
            return
        }
        player.team.DelPlayer(userID)
        player.team = nil
        team.Broadcast("user.state", map[string]interface{}{ "id": userID, "online": false, "members": getMembers(team) })
        player.SendMsg("game.exit", nil)
    })

    bind("dissolve", func(player *Player, data map[string]interface{}) {
        if player.userID != player.team.creator{
            return
        }
        player.BroadcastInTeam("game.exit", nil)

        var teamID = player.team.teamID
        player.team.Foreach(func (player *Player) {
            player.team = nil
        })

        DelTeam(teamID)
        player.mapObj.DelTeam(teamID)
    })

    bind("ping", func(player *Player, data map[string]interface{}) {
        player.SendMsg("game.pong", nil)
    })
}

func onSessionReq(sessionid int, msg string){
    var data map[string]interface{}
    if err := json.Unmarshal([]byte(msg), &data); err != nil {
        log.Println(err)
        return
    }
    name := data["n"].(string)
    d,ok := data["d"].(map[string]interface{})
    if !ok{
        d = nil
    }
    data = d
    fn := gMsgCallBack[name]
    log.Println(name)

    if fn != nil{
        player := GetPlayerBySessonId(sessionid)
        if player == nil{
            if name == "login"{
                if util.IsValid(data["sign"].(string)) == false && data["sign"].(string) != "notoken"{
                    log.Printf("sign error")
                    return
                }
                player = NewPlayer(sessionid)
            }else{
                return
            }
        }
        fn(player, data)
    }
}

func onSessionOffline(sessionid int){
    log.Printf("onSessionOffline", sessionid)
    player := GetPlayerBySessonId(sessionid)
    if player == nil{
        return
    }

    if player.team != nil {
        mapObj := player.mapObj
        DelPlayer(player.userID)    
        mapObj.DelPlayer(player.userID)
        player.mapObj = nil
    }else{
        player.online = false
        player.NotifyOtherInTeam("user.offline", map[string]interface{}{"id":player.userID})
    }
}

func main() {
    log.Println("start worker...")

    db.InitDB()

    loadJson()

    bindHandler()

    worker.Open("localhost:43210", onSessionReq, onSessionOffline)
    
    log.Println("input anykey to exit...")
    input := bufio.NewReader(os.Stdin)
    input.ReadString('\n')
}