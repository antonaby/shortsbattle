import { defineStore } from "pinia";
import { Centrifuge, Subscription } from "centrifuge";
import { useGameStore } from "./gameStore";

interface CentrifugeStoreState {
  client: Centrifuge | null;
  sub: Subscription | null;
  connected: boolean;
}

export const useCentrifugeStore = defineStore("centrifuge", {
  state: (): CentrifugeStoreState => ({
    client: null,
    sub: null,
    connected: false,
  }),
  actions: {
    connect() {
      const cf = new Centrifuge(
        `${import.meta.env.VITE_WS_BASE_URL}/api/v1/games/updates`
      );

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
    subscribe(gameInstanceId: number) {
      if (this.client == null) {
        return;
      }

      const sub = this.client.newSubscription(`game_${gameInstanceId}`);
      sub.on("publication", (ctx) => {
        const gameStore = useGameStore()
        gameStore.updateGameStatus(ctx.data.status)
      });
      sub.subscribe();

      this.sub = sub;
    },
    unsubsribe() {
      if (this.sub == null) {
        return;
      }

      this.sub.unsubscribe();
      this.sub = null;
    },
  },
});
