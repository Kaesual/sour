declare module 'cbor-js'

declare module 'url:*' {
  export default string
}

type PreloadFile = {
  filename: string
  start: number
  end: number
  audio?: number
}
type PreloadNode = {
  name: string
  files: PreloadFile[]
}

// Lets us use the Module API in a type safe way
type ModuleType = {
  HEAPU8: Uint8Array
  _free: (pointer: number) => void
  _malloc: (length: number) => number
  canvas: HTMLCanvasElement | null
  desiredHeight: number
  desiredWidth: number
  monitorRunDependencies: (left: number) => void
  postLoadWorld: () => void
  postRun: Array<() => void>
  preRun: Array<() => void>
  preInit: Array<() => void>
  onGameReady: () => void
  print: (text: string) => void
  registerNode: (node: PreloadNode) => void
  removeRunDependency: (file: string) => void
  run: () => void
  setCanvasSize: ((width: number, height: number) => void) | null
  setStatus: (text: string) => void
  tweakDetail: () => void
  running: boolean

  calledRun: boolean
  FS_createPath: (...path: Array<string, boolean>) => void
  FS_createPreloadedFile: (
    parent: string,
    name: Maybe<string>,
    url: string | Uint8Array,
    canRead: boolean,
    canWrite: boolean,
    onload: () => void,
    onerror: () => void,
    dontCreateFile: boolean,
    canOwn: boolean,
    preFinish?: () => void
  ) => void
  FS_createDataFile: (
    parent: string,
    name: string,
    something: string,
    canRead: boolean,
    canWrite: boolean,
    canOwn: boolean
  ) => void
  addRunDependency: (dependency: string) => void
  socket: (addr: string, port: number) => any

  onConnect: (hostname: string, port: number) => void
  onDisconnect: () => void
  onClientJoin: (name: string) => void

  // Run a JavaScript command and return a pointer to its result
  // This is NOT the same thing as emscripten_run_script_string
  interop: (script: string) => number

  assets: {
    // assets has its own hook
    onConnect: () => void
    missingSound: (path: string, msg: number) => void
    missingTexture: (path: string, msg: number) => void
    missingModel: (name: string, msg: number) => void
    loadRandomMap: () => void
    loadWorld: (map: string) => void
    receiveMap: (map: string, oldMap: string) => void
    installMod: (name: string) => void
    getModProperty: (id: string, property: string) => string
    modsToURL: () => void
  }

  gameState: {
    playerState: (
      health: number,
      maxHealth: number,
      gunSelect: number,
      shotgunAmmo: number,
      chainggunAmmo: number,
      rocketAmmo: number,
      rifleAmmo: number,
      grenadeAmmo: number,
      pistolAmmo: number,
    ) => void
  },

  loadedMap: (name: string) => void
  onLocalDisconnect: () => void

  discord: {
    login: () => void
    copyKey: () => void
    regenKey: () => void
    logout: () => void
  }

  cluster: {
    createGame: (preset: string, mode: string) => void
    connect: (name: string, password: string) => void
    send: (channel: number, dataPtr: number, dataLength: number) => void
    receive: (dataPtr: number, dataLengthPtr: number) => void
    disconnect: () => void
  }
}
declare const Module: ModuleType
declare type Maybe<T> = T | null | undefined

declare const lengthBytesUTF8 = (s: string) => number
declare const stringToUTF8 = (s: string, a: number, bytes: number) => number

declare const FS: {
  unlink: (file: string) => void
  lookupPath: (file: string) => Maybe<any>
}

declare const WASM_PROMISE: Promise<void>

declare const INJECTED_SOUR_CONFIG: Maybe<any>

type BananaBreadType = {
  conoutf: (level: number, message: string) => void
  execute: (command: string) => void
  mousemove: (dx: number, dy: number) => void
  isInMenu: () => number
  click: (x: number, y: number) => void
  loadWorld: (map: string, cmap?: string) => void
  setLoading: (value: boolean) => void
  renderprogress: (progress: number, text: string) => void
  injectServer: (
    host: string,
    port: number,
    infoPointer: number,
    infoLength: number
  ) => void
}
declare const BananaBread: BananaBreadType
declare let shouldRunNow: boolean
declare let calledRun: boolean
