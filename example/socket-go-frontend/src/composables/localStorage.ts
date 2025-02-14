const keys = {
    NAME: 'name'
} as const

export const nameStore = {
    get: () => localStorage.getItem(keys.NAME),
    set: (value: string) => localStorage.setItem(keys.NAME, value)
}