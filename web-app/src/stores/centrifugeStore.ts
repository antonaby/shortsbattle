import { defineStore } from "pinia";
import { Centrifuge, Subscription } from "centrifuge";

interface GameUpdate {
  id: number;
  status: string;
}

interface CentrifugeStoreState {
  client: Centrifuge | null;
  sub: Subscription | null;
  connected: boolean;
  lastGameUpdate: GameUpdate | null;
}

export const useCentrifugeStore = defineStore("centrifuge", {
  state: (): CentrifugeStoreState => ({
    client: null,
    sub: null,
    connected: false,
    lastGameUpdate: null,
  }),
  actions: {
    connect(gameInstanceId: number) {
      const cf = new Centrifuge("ws://localhost:8080/api/v1/join");

      const sub = cf.newSubscription(`game_${gameInstanceId}`);
      sub.on("publication", (ctx) => {
        this.lastGameUpdate = ctx.data;
      });
      sub.subscribe();
      this.sub = sub;

      cf.on("connected", () => {
        this.connected = true;
      });
      cf.on("disconnected", () => {
        this.connected = false;
      });
      cf.connect();

      this.client = cf;
    },
    disconnect() {
      this.sub?.unsubscribe();
      this.sub = null;
      this.client?.disconnect();
      this.client = null;
    },
  },
});
