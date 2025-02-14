import {SocketGo} from "socket-go";

type EmitData = {
    "set_name": string
    "join_room": string
    "send_massage": string
}
type OnData = {
    "echo": string
    "send_massage": {
        name: string
        message: string
        time: string
    }
    "join_room": string
}

export type Socket = SocketGo<EmitData, OnData>

let ws: Socket | null = null

export const useSocket = () => {
    if (!ws) {
        ws = new SocketGo(`ws://localhost:8080/ws`)
    }
    return ws
}