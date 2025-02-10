type handlerFunc<T = any> = (data?: T) => void

const sleep = (ms: number = 1000) =>
    new Promise((resolve) => setTimeout(resolve, ms))

class SocketGo<
    EmitData extends Record<string, any>,
    OnData extends Record<string, any>,
> {
    private _ws: WebSocket
    private _isConnected: boolean = false
    private _handler: Record<string, handlerFunc> = {}
    private _onConnectHandler: handlerFunc | null = null
    private _onDisconnectHandler: handlerFunc | null = null

    constructor(url: string | URL) {
        this._ws = new WebSocket(url)

        this.init(url)
    }

    private init(url: string | URL) {
        this._ws.onmessage = (e) => {
            const obj = JSON.parse(e.data)
            const event = obj['event']
            const data = obj['data']

            if (event && this._handler[event]) {
                this._handler[event](data)
            }
        }

        this._ws.onclose = async () => {
            if (this._isConnected) {
                this._onDisconnectHandler?.()
            }

            this._isConnected = false
            await sleep()
            this._ws = new WebSocket(url)
            this.init(url)
        }

        this._ws.onopen = () => {
            this._isConnected = true
            this._onConnectHandler?.()
            this.keepAlive(this._ws)
        }
    }

    private keepAlive(ws: WebSocket) {
        sleep(5000).then(() => {
            if (this._ws != ws) return
            this.emit("ping")
            this.keepAlive(ws)
        })
    }

    // 傳出事件
    // 建議Socket 只用於接收 而不進行傳送資料
    emit<K extends keyof EmitData>(event: K, data?: EmitData[K]) {
        this._ws.send(JSON.stringify({event, data}))
    }

    // 監聽事件
    on<K extends keyof OnData>(
        event: K,
        handler: handlerFunc<OnData[K] | undefined>
    ) {
        if (event === 'ping') {
            throw new Error(`"ping" is a reserved keyword.`)
        }

        if (typeof handler === 'function') {
            this._handler[event as string] = handler
        } else {
            throw new Error(`handler is not a function`)
        }
    }

    // 取消監聽事件
    off<K extends keyof OnData>(event: K) {
        delete this._handler[event as string]
    }

    // Websocket onopen 觘發
    onConnect(handler: handlerFunc) {
        this._onConnectHandler = handler
    }

    // 只有在成功建立連線後才會觸發
    onDisconnect(handler: handlerFunc<void>) {
        this._onDisconnectHandler = handler
    }
}

export {SocketGo}
export default SocketGo
