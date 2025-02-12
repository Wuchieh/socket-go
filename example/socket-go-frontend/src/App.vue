<template>
  <div class="container">
    <dialog ref="dialogRef">
      <div>
        <form @submit.prevent="onSubmitName">
          <s-input label="請輸入暱稱" v-model="name" class="mb-2"/>
          <s-btn>送出</s-btn>
        </form>
      </div>
    </dialog>
    <div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {SocketGo} from 'socket-go'
import {onMounted, ref} from "vue";
import SInput from "./components/SInput.vue";
import SBtn from "./components/SBtn.vue";

const ws = new SocketGo("ws://localhost:8080/ws")
const dialogRef = ref<HTMLDialogElement | null>(null)
const name = ref("")
const onSubmitName = () => {
  ws.emit("echo",name.value)
}

onMounted(() => {
  dialogRef.value?.showModal()
})
</script>

<style scoped>
</style>