type handlerFunc<T = any> = (data: T) => void;

class SocketGo<EmitData extends Record<string, any>, OnData extends Record<string, any>> {
    ws: WebSocket;
    handler: Record<string, handlerFunc>;

    constructor(url: string | URL) {
        this.ws = new WebSocket(url);
        this.handler = {};

        this.ws.onmessage = (e) => {
            const obj = JSON.parse(e.data);
            const event = obj['event'];
            const data = obj['data'];

            if (event && this.handler[event]) {
                this.handler[event](data);
            }
        };
    }

    emit<K extends keyof EmitData>(event: K, data?: EmitData[K]) {
        this.ws.send(JSON.stringify({event, data}));
    }

    on<K extends keyof OnData>(event: K, handler: handlerFunc<OnData[K] | undefined>) {
        if (typeof handler === 'function') {
            this.handler[event as string] = handler;
        }
    }
}

export {SocketGo};
