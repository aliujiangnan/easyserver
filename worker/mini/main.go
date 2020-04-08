package main

import (
    "bufio"
    "log"
    "os"
    "math"
    "strconv"
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

func gameBegin(player *Player, team *Team){
    players := make([]map[string]interface{}, 0)
    team.Foreach(func(playerInTeam *Player){
        info := map[string]interface{}{
            "id": playerInTeam.userID,
            "user": playerInTeam.userName,
            "nick": playerInTeam.userName,
            "online": playerInTeam.online,
            "gender": playerInTeam.gender,
            "avatar": playerInTeam.avatar,
            "score": playerInTeam.score,
            "level": playerInTeam.level,
        }
        players = append(players, info)
    })
    if (team.hasRobot){
        info := map[string]interface{}{
            "id": player.userID + 1,
            "user": strconv.Itoa(int(time.Now().Unix() * 1000)),
            "nick": strconv.Itoa(int(time.Now().Unix() * 1000)),
            "online": true,
            "gender": 0,
            "avatar": "http://sandbox-avatar.boochat.cn/2018/01/20/02/3/0042758573.gif",
            "score": 0,
            "level": 1,
        }
        players = append(players, info)
    }
    var data = map[string]interface{}{
        "t":1,
        "s":1,
        "id":team.teamID,
        "players":players,
        "robot":util.If(team.hasRobot, 1 , 0),
    }
    player.BroadcastInTeam("game.begin", data)
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
    gameType := int(data["game"].(float64))
    db.GetScore(userID, gameType, func(score int){
        player.score = score
        player.score = util.If(player.score < 0 , 0 , player.score).(int)
        db.GetExp(userID, gameType, func(exp int){
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

    bind("match", func(player *Player, data map[string]interface{}){
        log.Printf("matching")
        var team = player.mapObj.MatchTeam(player, int(data["type"].(float64)))
        log.Printf("teamid, %d, num %d", team.teamID, team.GetMemberNum())
        if team.GetMemberNum() == MAX_NUM_EACH_TEAM {
            gameBegin(player, team)
        }
        util.SetTimeout(func(){
            if team.GetMemberNum() == MAX_NUM_EACH_TEAM || team.hasRobot {
                return
            }
            team.hasRobot = true
            gameBegin(player, team)
        },15000)
    })


    bind("msg",func(player *Player, data map[string]interface{}){
        player.NotifyOtherInTeam("player.msg", data)
    })


    bind("bcast", func(player *Player, data map[string]interface{}){
        player.BroadcastInTeam("player.msg", data)
    })


    bind("emoji", func(player *Player, data map[string]interface{}){
        player.NotifyOtherInTeam("player.emoji", data)
    })

    bind("again", func(player *Player, data map[string]interface{}){
        player.NotifyOtherInTeam("player.again", data)
    })

    bind("accept", func(player *Player, data map[string]interface{}){
        gameBegin(player, player.team)
    })

    bind("dissolve",func(player *Player, data map[string]interface{}){
        player.BroadcastInTeam("game.dissolve", nil)
        DelTeam(player.team.teamID)
        player.mapObj.DelTeam(player.team.teamID)
        player.team = nil
    })

    bind("gameover",func(player *Player, data map[string]interface{}){
        rst := int(data["rst"].(float64))

        player.team.Foreach(func(playerInTeam *Player){
            result := 0
            if (playerInTeam.userID == player.userID){
                result = rst
            }else if (rst == 2){
                result = rst
            }else{
                result = util.If(rst == 0 , 1 , 0).(int)
            }
            expAdd := util.If(result == 0 , 20 , 40).(int)
            scoreAdd := util.If(result == 0, -10, 20).(int)
            beforeNext := int(40 * math.Pow(2, float64(playerInTeam.level))) - int(40 * math.Pow(2, float64(playerInTeam.level - 1)))
            beforeLast := util.If(playerInTeam.level == 1 , 0 , int(40 * math.Pow(2, float64(playerInTeam.level - 1))) - int(40 * math.Pow(2, float64(player.level - 2)))).(int)
            beforeExp := playerInTeam.exp - beforeLast
            curExp := playerInTeam.exp + expAdd
            curLv := getLevel(curExp)
            curScore := playerInTeam.score + scoreAdd
            curScore = util.If(curScore < 0, 0, curScore).(int)
            nextExp := int(40 * math.Pow(2, float64(curLv))) - int(40 * math.Pow(2, float64(curLv - 1)))
            lastExp := util.If(curLv == 1, 0, int(40 * math.Pow(2, float64(curLv - 1))) - int(40 * math.Pow(2, float64(curLv - 2)))).(int)

            data := map[string]interface{}{
                "curlv":curLv,
                "curscore":curScore,
                "befoexp":beforeExp,
                "befonext":beforeNext,
                "curexp":curExp - lastExp,
                "nextexp":nextExp,
                "scoreadd": util.If(scoreAdd >= 0 , "+" , "").(string) + strconv.Itoa(scoreAdd),
                "expadd":"+" + strconv.Itoa(expAdd),
            }

            player.score = curScore
            player.exp = curExp
            player.level = curLv
            db.UpdateScore(player.userID, player.team.selfType, player.score)
            db.AddExp(player.userID, player.team.selfType, expAdd)
            playerInTeam.SendMsg("game.over", map[string]interface{}{"rst":result, "data":data})

            if (playerInTeam.online == false){
                DelPlayer(playerInTeam.userID)
                playerInTeam.team.DelPlayer(playerInTeam.userID)
                mapObj := playerInTeam.mapObj
                mapObj.DelPlayer(playerInTeam.userID)
                playerInTeam.mapObj = nil
            }
        })

        if (player.team.GetMemberNum() < MAX_NUM_EACH_TEAM){
            DelTeam(player.team.teamID)
            player.mapObj.DelTeam(player.team.teamID)
            player.team = nil
        }
    })

    bind("ping", func(player *Player, data map[string]interface{}){
        player.SendMsg("game.pong",nil)
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

    bindHandler()

    worker.Open("localhost:43210", onSessionReq, onSessionOffline)
    
    log.Println("input anykey to exit...")
    input := bufio.NewReader(os.Stdin)
    input.ReadString('\n')
}