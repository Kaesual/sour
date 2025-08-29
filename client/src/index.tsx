import styled from '@emotion/styled'
import { useResizeDetector } from 'react-resize-detector'
import start from './unsafe-startup'
import CBOR from 'cbor-js'
import * as React from 'react'
import * as R from 'ramda'
import ReactDOM from 'react-dom'
import {
  Center,
  ChakraProvider,
  Button,
  extendTheme,
  Flex,
  Box,
  VStack,
  Heading,
  Spacer,
} from '@chakra-ui/react'

import type { ThemeConfig } from '@chakra-ui/react'

import type { GameState, PlayerState } from './types'
import type {
  ClientAuthMessage,
  ServerMessage,
  SocketMessage,
  CommandMessage,
  PacketMessage,
} from './protocol'
import { GameStateType, WeaponType } from './types'
import { MessageType, ENetEventType } from './protocol'
import StatusOverlay from './Loading'
import NAMES from './names'
import useAssets, { getInstalledMods, mountFile } from './assets/hook'
import useAuth, {
  DISCORD_CODE,
  renderDiscordHeader,
  renderDiscordButton,
} from './discord'
import { CubeMessageType } from './game'
import * as cube from './game'
import MobileControls from './MobileControls'
import FileDropper from './FileDropper'

import type { PromiseSet } from './utils'
import { CONFIG } from './config'
import { breakPromise, BROWSER } from './utils'
import * as log from './logging'

import { LoadRequestType } from './assets/types'

start()

const colors = {
  brand: {
    900: '#1a365d',
    800: '#153e75',
    700: '#2a69ac',
  },
}

const config: ThemeConfig = {
  initialColorMode: 'dark',
  useSystemColorMode: false,
}

const theme = extendTheme({ colors, config })

const OuterContainer = styled.div`
  touch-action: none;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
`

const GameContainer = styled.div`
  width: 100%;
  height: 100%;
  position: absolute;
  z-index: 0;
`

const LoadingContainer = styled.div`
  width: 100%;
  height: 100%;
  position: absolute;
  z-index: 1;
`

const DropTarget = styled.div`
  width: 100%;
  height: 100%;
  position: absolute;
  pointer-events: none;
  background: var(--chakra-colors-yellow-500);
  z-index: 2;
`

const pushURLState = (url: string) => {
  const {
    location: { search: params },
  } = window

  // We only want to keep 'mods', everything else goes
  const parsedParams = new URLSearchParams(params)
  const newParams = new URLSearchParams()
  if (parsedParams.has('mods')) {
    const value = parsedParams.get('mods')
    if (value != null) {
      newParams.set('mods', value)
    }
  }

  window.history.pushState({}, '', `${url}${newParams.toString()}`)
}

const clearURLState = () => pushURLState('/')

export type CommandRequest = {
  id: number
  promiseSet: PromiseSet<string>
}

const DEBUG = false
const DELAY_AFTER_LOAD: CubeMessageType[] = [
  CubeMessageType.N_ITEMLIST,
  CubeMessageType.N_SPAWN,
]

const SERVER_URL_REGEX = /#\/server\/([\w.]+)\/?(\d+)?/
const MAP_URL_REGEX = /#\/map\/(\w+)/
const DEMO_URL_REGEX = /#\/demo\/(\w+)/

let loadedMods: string[] = []
let failedMods: string[] = []

async function playDemoURL(url: string, reference: string) {
  log.info(`loading demo ${reference}`)
  try {
    const response = await fetch(url)
    if (response.status === 404) {
      log.error(`demo ${reference} not found`)
      return
    }
    if (response.status !== 200) {
      throw Error('failed to fetch')
    }

    const data = await response.arrayBuffer()
    const demoName = `/demo/${reference}.dmo`
    await mountFile(demoName, new Uint8Array(data))
    BananaBread.execute(`demo ${reference}.dmo`)
  } catch (e) {
    log.error('failed to play demo')
  }
}

function App() {
  const [state, setState] = React.useState<GameState>({
    type: GameStateType.PageLoading,
  })
  const { width, height, ref: containerRef } = useResizeDetector()

  const canvasRef = React.useRef<HTMLCanvasElement>(null)
  const wsRef = React.useRef<WebSocket>()
  const wsQueue = React.useRef<ArrayBuffer[]>([])

  const send = React.useCallback((data: ArrayBuffer) => {
    const { current: ws } = wsRef
    const { current: queue } = wsQueue
    if (ws == null || ws.readyState !== WebSocket.OPEN) {
      wsQueue.current = [...queue, data]
      return
    }
    ws.send(data)
  }, [])

  const sendAuthMessage = React.useCallback((message: ClientAuthMessage) => {
    send(CBOR.encode(message))
  }, [])

  const {
    loadAsset,
    loadAssetProgress,
    getMod,
    onReady: onReadyAssets,
  } = useAssets(setState)
  const {
    state: authState,
    receiveMessage: receiveAuthMessage,
    initialize: initializeDiscord,
  } = useAuth(sendAuthMessage)

  const [playerState, setPlayerState] = React.useState<PlayerState>({
    health: 0,
    maxHealth: 0,
    weapon: WeaponType.Pistol,
    ammo: {
      [WeaponType.Saw]: 0,
      [WeaponType.Shotgun]: 0,
      [WeaponType.Chaingun]: 0,
      [WeaponType.Rocket]: 0,
      [WeaponType.Rifle]: 0,
      [WeaponType.Grenade]: 0,
      [WeaponType.Pistol]: 0,
    },
  })

  React.useEffect(() => {
    Module.gameState = {
      playerState: (
        health: number,
        maxHealth: number,
        gunSelect: number,
        shotgunAmmo: number,
        chainggunAmmo: number,
        rocketAmmo: number,
        rifleAmmo: number,
        grenadeAmmo: number,
        pistolAmmo: number
      ) => {
        setPlayerState({
          health,
          maxHealth,
          weapon: gunSelect,
          ammo: {
            [WeaponType.Saw]: 1,
            [WeaponType.Shotgun]: shotgunAmmo,
            [WeaponType.Chaingun]: chainggunAmmo,
            [WeaponType.Rocket]: rocketAmmo,
            [WeaponType.Rifle]: rifleAmmo,
            [WeaponType.Grenade]: grenadeAmmo,
            [WeaponType.Pistol]: pistolAmmo,
          },
        })
      },
    }
  }, [])

  React.useEffect(() => {
    ;(async () => {
      // Waits for the WASM file to be downloaded and memory to be initialized
      // This solves a race condition wherein we were mounting things to the
      // filesystem before the HEAP was defined
      await WASM_PROMISE

      // Not a mod but I'm lazy
      await loadAsset(LoadRequestType.Mod, 'environment')

      // Load the basic required data for the game
      await loadAsset(LoadRequestType.Mod, 'base')
      await loadAsset(LoadRequestType.Mod, 'fps')

      const {
        location: { search: params },
      } = window

      const parsedParams = new URLSearchParams(params)
      let mods: string[] = getInstalledMods()
      if (parsedParams.has('safemode')) {
        mods = []
      }
      if (parsedParams.has('mods')) {
        const urlMods = parsedParams.get('mods')
        if (urlMods != null) {
          mods = urlMods.split(',')
        }
      }

      if (mods.length > 0) {
        await Promise.all(
          R.map(async (id: string) => {
            const mod = getMod(id)
            if (mod == null) {
              failedMods = [...failedMods, id]
              return
            }

            const { name } = mod

            try {
              const layer = await loadAsset(LoadRequestType.Mod, id)
              if (layer == null) {
                failedMods = [...failedMods, name]
                return
              }
              loadedMods = [...loadedMods, name]
            } catch (e) {
              failedMods = [...failedMods, name]
            }
          }, mods)
        )
      }

      shouldRunNow = true
      calledRun = false
      Module.calledRun = false
      Module.run()
    })()

    Module.socket = (addr, port) => {
      const { protocol, host } = window.location
      const prefix = `${
        protocol === 'https:' ? 'wss://' : 'ws:/'
      }${host}/service/proxy/`

      return new WebSocket(
        addr === 'sour' ? prefix : `${prefix}u/${addr}:${port}`,
        ['binary']
      )
    }

    Module.print = (text) => {
      if (text === 'init: sdl') {
        setState({
          type: GameStateType.Running,
        })
      }

      // Randomly assign a new name if the user joins without one
      if (text === 'setting name to: unnamed') {
        const name = NAMES[Math.floor(Math.random() * NAMES.length)]
        BananaBread.execute(`name ${name}`)
      }

      if (text.startsWith('main loop blocker')) {
        return
      }

      console.log(text)
    }
  }, [])

  const setResolution = React.useCallback(
    (width: Maybe<number>, height: Maybe<number>) => {
      if (BROWSER.isMobile) {
        const ratio = window.devicePixelRatio || 1
        const {
          documentElement: { clientWidth, clientHeight },
        } = document
        const width = clientWidth * ratio
        const height = clientHeight * ratio
        const { canvas } = Module
        if (canvas == null) return

        if (Module.running) {
          if (BananaBread == null || BananaBread.execute == null) return
          BananaBread.execute(`screenres ${width * 2} ${height * 2}`)
        }

        canvas.style.setProperty('width', clientWidth + 'px', 'important')
        canvas.style.setProperty('height', clientHeight + 'px', 'important')
        canvas.width = width * 2
        canvas.height = height * 2
        return
      }

      if (width == null || height == null) return
      Module.desiredWidth = width
      Module.desiredHeight = height
      if (Module.setCanvasSize == null) return
      Module.setCanvasSize(width, height)
      if (BananaBread == null || BananaBread.execute == null) return
      BananaBread.execute(`screenres ${width} ${height}`)
    },
    []
  )

  React.useEffect(() => {
    setResolution(width, height)
  }, [width, height])

  React.useEffect(() => {
    if (state.type !== GameStateType.Ready) return

    // To show the server browser, add:
    //   guibutton "server browser.." "showgui servers"
    // 
    // Proxy setup is incomplete though at the moment, so
    // connecting won't work

    // Removed discord button for now
    // 
    //   newgui discord [
    //       guibutton "copy authkey command.." [js "Module.discord.copyKey()"]
    //       //guibutton "regenerate auth key.." [js "Module.discord.regenKey()"]
    //       guibutton "log out.." [js "Module.discord.logout()"]
    //   ]


    const menu = `
    newgui content [
        guibutton "mods.."  "showgui mods"
        guibutton "put mods in url.."  [js "Module.assets.modsToURL()"]
        guibutton "reload page.."  [js "window.location.reload()"]
    ]

    injectedmenu = [
        guilist [
          guiimage (concatword "packages/icons/" (playermodelicon) ".jpg") [chooseplayermodel] 1.15
          guistrut 0.25
          guilist [
              newname = (getname)
              guifield newname 15 [name $newname]
              guispring
              guilist [
                  guibutton (playermodelname) [chooseplayermodel] 0
                  guistrut 1
                  guiimage (getcrosshair) [showgui crosshair] 0.5
              ]
          ]
      ]
      ${renderDiscordHeader(authState)}
      guibar
      if (isconnected) [
          if (|| $editing (m_edit (getmode))) [
              guibutton "editing.." "showgui editing"
          ]
          guibutton "vote game mode / map.." "showgui gamemode"
          guibutton "switch team" [if (strcmp (getteam) "good") [team evil] [team good]]
          guibutton "toggle spectator" [spectator (! (isspectator (getclientnum)))] "spectator"
          guibutton "master.." [showgui master]
          guibutton "disconnect" "disconnect"         "exit"
          guibar
      ] [
          ${CONFIG.menuOptions}
          guibutton "join insta-dust2" "join insta-dust2"
          guibutton "join ffa-dust2" "join ffa-dust2"
          guibutton "join insta rotating maps" "join insta"
          guibutton "join ffa rotating maps" "join lobby"
          guibutton "create private game..." "creategame ffa"
      ]
      guibutton "random map.."  "map random"
      guibutton "content.." "showgui content"
      if ($fullscreen) [
          guibutton "exit fullscreen.." [fullscreen 0]
      ] [
          guibutton "enter fullscreen.." [fullscreen 1]
      ]
      ${renderDiscordButton(authState)}
      guibutton "options.."        "showgui options"
      guibutton "about.."          "showgui about"
    ]
    `
    BananaBread.execute(menu)
  }, [authState, state])

  React.useEffect(() => {
    // All commands in flight
    let commands: CommandRequest[] = []

    const [serverURL] = CONFIG.servers

    const { protocol, host } = window.location
    const ws = new WebSocket(
      `${protocol === 'https:' ? 'wss://' : 'ws:/'}${serverURL}`
    )
    ws.binaryType = 'arraybuffer'

    ws.onopen = () => {
      const { current: queue } = wsQueue
      if (queue == null) return
      for (const message of queue) {
        send(message)
      }
    }

    wsRef.current = ws

    const runCommand = async (command: string) => {
      const generate = (): number => Math.floor(Math.random() * 2048)

      let id: number = generate()

      // We don't want collisions and can't use a Symbol
      while (R.find((v) => v.id === id, commands) != null) {
        id = generate()
      }

      const promiseSet = breakPromise<string>()

      commands = [
        ...commands,
        {
          id,
          promiseSet,
        },
      ]

      const message: CommandMessage = {
        Op: MessageType.Command,
        Command: command,
        Id: id,
      }

      send(CBOR.encode(message))

      return promiseSet.promise
    }

    const injectServers = (servers: any) => {
      R.map((server) => {
        const { Host, Port, Info, Length } = server

        // Get data byte size, allocate memory on Emscripten heap, and get pointer
        const pointer = Module._malloc(Length)

        // Copy data to Emscripten heap (directly accessed from Module.HEAPU8)
        const dataHeap = new Uint8Array(Module.HEAPU8.buffer, pointer, Length)
        dataHeap.set(new Uint8Array(Info.buffer, Info.byteOffset, Length))

        // Call function and get result
        BananaBread.injectServer(Host, Port, pointer, Length)

        // Free memory
        Module._free(pointer)
      }, servers)
      BananaBread.execute('sortservers')
    }

    let serverEvents: SocketMessage[] = []
    let queuedEvents: SocketMessage[] = []
    let loadingWorld = false

    Module.running = false
    Module.postLoadWorld = function () {
      loadingWorld = false
      serverEvents = [...serverEvents, ...queuedEvents]
    }
    Module.addRunDependency = (file) => {
      console.log(`add ${file}`)
    }

    Module.interop = (command: string): number => {
      try {
        const result = window.eval(command)
        if (result == null) {
          return 0
        }
        const numBytes = lengthBytesUTF8(result) + 1
        const onHeap = Module._malloc(numBytes)
        stringToUTF8(result, onHeap, numBytes)
        // Sauer frees this on its own when it's done
        return onHeap
      } catch (e) {
        console.error(`failed to eval '${command}'`, e)
        return 0
      }
    }

    let remoteConnected: boolean = false

    let cachedServers: Maybe<any> = null
    Module.onGameReady = () => {
      onReadyAssets()
      Module.FS_createPath(`/`, 'packages', true, true)
      Module.FS_createPath(`/packages`, 'base', true, true)
      Module.FS_createPath(`/packages`, 'prefab', true, true)
      Module.FS_createPath(`/`, 'demo', true, true)

      if (BROWSER.isFirefox || BROWSER.isSafari) {
        BananaBread.execute('skipparticles 1')
        BananaBread.execute('glare 0')
      } else {
        BananaBread.execute('skipparticles 0')
      }

      if (!BROWSER.isMobile) {
        BananaBread.execute(`
              fullscreendesktop 1
        `)
      }

      if (BROWSER.isMobile) {
        BananaBread.execute(`
              // mobile screens are really dark
              lazyshader 0 "mobilegamma" (fsvs) (fsps [gl_FragColor.rgb = pow(color.rgb, vec3(1.0/1.8));])
              setpostfx mobilegamma
              forceplayermodels
              playermodel 1
              gui2d 1
              skyboxglare 0
              fullscreendesktop 0
              skipskybox 1
        `)
      }

      Module.running = true
      setResolution(null, null)
      setState({
        type: GameStateType.Ready,
      })

      if (cachedServers != null) {
        injectServers(cachedServers)
      }

      if (loadedMods.length > 0) {
        log.success(
          `loaded mods: ${R.join(
            ', ',
            loadedMods
          )}. if you experience issues, add ?safemode to the URL.`
        )
      }
      if (failedMods.length > 0) {
        log.error(`failed to load mods: ${R.join(', ', failedMods)}`)
      }

      const {
        location: { search: params, hash },
      } = window

      const serverDestination = SERVER_URL_REGEX.exec(hash)
      const mapDestination = MAP_URL_REGEX.exec(hash)
      if (serverDestination != null) {
        const [, hostname, port] = serverDestination
        if (port == null) {
          BananaBread.execute(`join ${hostname}`)
        } else {
          BananaBread.execute(`connect ${hostname} ${port}`)
        }
      } else if (mapDestination != null) {
        const [, mapId] = mapDestination
        BananaBread.execute(`map ${mapId}`)
      } else if (hash.startsWith('/demo/')) {
        // First check the fragment for a URL
        if (hash.length !== 0) {
          const url = hash.slice(1)
          remoteConnected = true
          playDemoURL(url, 'from-url')
        } else {
          const demoDestination = DEMO_URL_REGEX.exec(hash)
          if (demoDestination != null) {
            const [, demoId] = demoDestination
            const [serverURL] = CONFIG.servers
            const { protocol } = window.location
            const demoURL = `${protocol}//${serverURL}api/demo/${demoId}`
            remoteConnected = true
            playDemoURL(demoURL, demoId)
          } else {
            pushURLState('#')
          }
        }
      } else {
        // It should not be anything else
        pushURLState('#')
      }

      const parsedParams = new URLSearchParams(params)

      initializeDiscord(
        parsedParams.has('code') ? parsedParams.get('code') : null
      )

      if (parsedParams.has('cmd')) {
        const cmd = parsedParams.get('cmd')
        if (cmd == null) return
        setTimeout(() => BananaBread.execute(cmd), 0)
      }

      if (parsedParams.has('safemode')) {
        log.info(
          'you are in safe mode. no mods were loaded, but you can disable them in the mod menu.'
        )
      }
    }

    const updateServerURL = (name: string, port: number) => {
      // Sour server
      if (port === 0) {
        pushURLState(`#/server/${name}`)
        return
      }

      pushURLState(`#/server/${name}/${port}`)
    }

    Module.onConnect = () => {}
    Module.onDisconnect = () => {
      remoteConnected = false
      clearURLState()
    }

    Module.loadedMap = (name: string) => {
      if (remoteConnected) return
      pushURLState(`#/map/${name}`)
    }

    let lastPointer: number = 0
    let lastPointerLength: number = 0

    // Only allocate memory if we really need to
    const malloc = (size: number) => {
      // reduce, reuse, recycle
      if (size <= lastPointerLength) {
        return lastPointer
      }

      if (lastPointer !== 0) {
        Module._free(lastPointer)
      }

      const pointer = Module._malloc(size)
      lastPointer = pointer
      lastPointerLength = size
      return pointer
    }

    Module.onLocalDisconnect = () => {
      clearURLState()
    }

    Module.cluster = {
      createGame: (preset: string, mode: string) => {
        log.info('creating private game...')
        ;(async () => {
          try {
            console.log(`creategame ${preset} ${mode}`)
            const result = await runCommand(`creategame ${preset} ${mode}`)
            log.success('created game!')
          } catch (e) {
            log.error(`failed to create private game: ${e}`)
          }
        })()
      },
      connect: (name: string, password: string) => {
        const Target = name.length === 0 ? 'lobby' : name
        send(
          CBOR.encode({
            Op: MessageType.Connect,
            Target,
          })
        )
      },
      send: (channel: number, dataPtr: number, dataLength: number) => {
        const packet = new Uint8Array(dataLength)
        packet.set(new Uint8Array(Module.HEAPU8.buffer, dataPtr, dataLength))
        if (DEBUG) {
          const p = cube.newPacket(packet)
          const msgType = cube.getInt(p)
          if (msgType != null) {
            console.log(
              `%c client -> server type=${CubeMessageType[msgType]}`,
              'background-color: blue; color: white'
            )
          }
        }
        send(
          CBOR.encode({
            Op: MessageType.Packet,
            Channel: channel,
            Data: packet,
            Length: dataLength,
          })
        )
      },
      receive: (dataPtr: number, dataLengthPtr: number) => {
        const view = new DataView(Module.HEAPU8.buffer)

        const message = serverEvents.shift()
        if (message == null) {
          return 0
        }

        if (message.Op === MessageType.ServerConnected) {
          // Layout:
          // 2: Event
          const frameLength = 2
          const pointer = malloc(frameLength)
          view.setUint16(pointer, ENetEventType.Connect, true)
          return pointer
        }

        if (message.Op === MessageType.ServerDisconnected) {
          const { Reason } = message

          // Layout:
          // 2: Event
          // 2: Reason
          const frameLength = 2 + 2
          const pointer = malloc(frameLength)
          view.setUint16(pointer, ENetEventType.Disconnect, true)
          view.setUint16(pointer + 2, Reason, true)
          return pointer
        }

        const { Channel, Data, Length } = message

        // Layout:
        // 2: Event
        // 2: Channel
        // 4: Length
        // N: Data
        const frameLength = 2 + 2 + 4 + Length
        const pointer = malloc(frameLength)

        // sourEvent
        view.setUint16(pointer, ENetEventType.Receive, true)
        // sourChannel
        view.setUint16(pointer + 2, Channel, true)
        // dataLength
        view.setUint32(pointer + 4, Length, true)

        // Copy in from data
        const dataHeap = new Uint8Array(
          Module.HEAPU8.buffer,
          pointer + 8,
          Length
        )
        dataHeap.set(new Uint8Array(Data.buffer, Data.byteOffset, Length))

        return pointer
      },
      disconnect: () => {
        clearURLState()
        send(
          CBOR.encode({
            Op: MessageType.Disconnect,
          })
        )
      },
    }

    ws.onmessage = (evt) => {
      const serverMessage: ServerMessage = CBOR.decode(evt.data)

      if (serverMessage.Op === MessageType.Info) {
        const { Cluster, Master } = serverMessage
        const combined = [...(Master || []), ...(Cluster || [])]

        if (
          BananaBread == null ||
          BananaBread.execute == null ||
          BananaBread.injectServer == null
        ) {
          cachedServers = combined
          return
        }

        injectServers(combined)
        return
      }

      if (serverMessage.Op === MessageType.ServerConnected) {
        remoteConnected = true
        const { Server, Internal, Owned } = serverMessage
        if (Internal) {
          clearURLState()
        } else {
          updateServerURL(Server, 0)
          if (Owned) {
            navigator.clipboard.writeText(location.href)
          }
        }
        // intentional fallthrough
      }

      if (serverMessage.Op === MessageType.ServerResponse) {
        const { Id, Response, Success } = serverMessage
        const request = R.find(({ id: otherId }) => Id === otherId, commands)
        if (request == null) return

        const {
          promiseSet: { resolve, reject },
        } = request

        if (Success) {
          resolve(Response)
        } else {
          reject(new Error(Response))
        }

        commands = R.filter(({ id: otherId }) => Id !== otherId, commands)
        return
      }

      if (serverMessage.Op === MessageType.Packet) {
        const packet = cube.newPacket(serverMessage.Data)
        const msgType = cube.getInt(packet)

        if (msgType === CubeMessageType.N_MAPCHANGE) {
          loadingWorld = true
          serverEvents.push(serverMessage)
          return
        }

        if (msgType != null) {
          if (DEBUG) {
            console.log(
              `%c server -> client type=${CubeMessageType[msgType]}`,
              'background-color: green; color: white'
            )
          }
          if (loadingWorld && !DELAY_AFTER_LOAD.includes(msgType)) {
            serverEvents.push(serverMessage)
            return
          }
        }
      }

      if (serverMessage.Op === MessageType.AuthSucceeded) {
        receiveAuthMessage(serverMessage)
        return
      }

      if (serverMessage.Op === MessageType.Chat) {
        log.chat(serverMessage.Message)
        return
      }

      if (serverMessage.Op === MessageType.AuthFailed) {
        receiveAuthMessage(serverMessage)
        return
      }

      if (loadingWorld) {
        queuedEvents.push(serverMessage)
      } else {
        serverEvents.push(serverMessage)
      }
    }
  }, [])

  React.useLayoutEffect(() => {
    const { canvas } = Module
    if (canvas == null) return

    // As a default initial behavior, pop up an alert when webgl context is lost. To make your
    // application robust, you may want to override this behavior before shipping!
    // See http://www.khronos.org/registry/webgl/specs/latest/1.0/#5.15.2
    canvas.addEventListener(
      'webglcontextlost',
      function (e) {
        alert('WebGL context lost. You will need to reload the page.')
        e.preventDefault()
      },
      false
    )

    //canvas.addEventListener('click', function () {
    //canvas.requestPointerLock()
    //})

    return
  }, [])

  return (
    <OuterContainer ref={containerRef}>
      <GameContainer>
        <canvas
          className="game"
          style={{ opacity: state.type !== GameStateType.Ready ? 0 : 1 }}
          id="canvas"
          tabIndex={0}
          ref={(canvas) => {
            if (canvas != null) {
              // This is a bug in mobile Safari where Reader holds on to canvas refs
              // https://gist.github.com/eugeneware/bb69a5bce8c8c48178845429435458e2#file-safari_reader-js-L1937
              // @ts-ignore
              canvas._evaluatedForTextContent = true
              // @ts-ignore
              canvas._cachedElementBoundingRect = {}
              // Ensure the canvas is focusable for keyboard events
              // (SDL with Emscripten binds keyboard to the target element)
              // eslint-disable-next-line no-param-reassign
              canvas.tabIndex = 0
            }
            Module.canvas = canvas
          }}
          onMouseDown={(_e: React.MouseEvent<HTMLCanvasElement>) => {
            // Focus the canvas so key events are delivered here
            // (important when SDL binds to #canvas)
            // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
            const el = document.getElementById('canvas') as HTMLCanvasElement | null
            if (el) el.focus()
          }}
          onContextMenu={(event) => event.preventDefault()}
        ></canvas>
        {BROWSER.isMobile && (
          <MobileControls
            playerState={playerState}
            isRunning={state.type === GameStateType.Ready}
          />
        )}
      </GameContainer>
      {state.type !== GameStateType.Ready && (
        <LoadingContainer>
          <Box w="100%" h="100%">
            <StatusOverlay state={state} />
          </Box>
        </LoadingContainer>
      )}
      <FileDropper>
        <DropTarget>
          <Flex align="center" justify="center">
            <VStack paddingTop="20%">
              <Heading>Drop a demo file here...</Heading>
            </VStack>
          </Flex>
        </DropTarget>
      </FileDropper>
    </OuterContainer>
  )
}

ReactDOM.render(
  <ChakraProvider theme={theme}>
    <App />
  </ChakraProvider>,
  document.getElementById('root')
)
