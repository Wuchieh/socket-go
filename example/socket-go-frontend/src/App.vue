<template>
  <div>
    <input type="text" v-model="message">
    <button @click="onClickEmit('echo')">echo</button>
    <button @click="onClickEmit('chat')">chat</button>
    <button @click="onClickJoinRoom">join room</button>
    <button @click="onClickBind">bind</button>
    <div>
      <div v-for="v in messages">
        {{ v }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import SocketGo from "./assets/socket-go";
import {onMounted, ref} from "vue";

type socket = SocketGo<any, any>

const message = ref('')
const messages = ref<string[]>([])
const socket = ref<socket | null>(null);
const onClickJoinRoom = () => {
  socket.value?.emit('join')
}

const onClickEmit = (e: string) => {
  const m = message.value
  socket.value?.emit(e, m)
}

const onClickBind = () => {
  socket.value?.emit('bind', {username: 'test'})
  socket.value?.emit('bind', 123)
  socket.value?.emit('bind', '123')
  socket.value?.emit('bind', 123.123)
  socket.value?.emit('bind')
}

const setupSocket = () => {
  const _socket = new SocketGo("ws://localhost:8080/ws");

  _socket.on('message', (msg: string) => {
    messages.value.push(msg);
  })

  socket.value = _socket
}

onMounted(() => {
  setupSocket()
})
</script>

<style scoped>

</style>