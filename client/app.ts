/** SOUNDBOARD CLIENT */
const log = (message: string) => console.log(`> ${message}`);
const error = (message: string) => console.error(`> ERROR: ${message}`);

type PlaySound = {
    name: string;
    start: number;
    end: number;
};

const soundboardClient = {
    _ws: null as WebSocket | null,
    _guildId: '',
    _channelId: '',

    init(guildId: string, channelId: string) {
        soundboardClient._guildId = guildId;
        soundboardClient._channelId = channelId;
    },

    _connect(): Promise<void> {
        const ws = soundboardClient._ws;
        // already open
        if (ws && ws?.readyState === WebSocket.OPEN) return Promise.resolve();

        // guildid/channelid not set
        if (!soundboardClient._guildId || !soundboardClient._channelId) {
            return Promise.reject('Guild id and/or channel id not set.');
        }

        return new Promise((resolve, reject) => {
            // init
            const ws = new WebSocket(
                `/ws?guild=${soundboardClient._guildId}&channel=${soundboardClient._channelId}`
            );
            soundboardClient._ws = ws;
            ws.onopen = (e) => {
                log('WS connection established');
                resolve();
            };
            ws.onclose = (e) => {
                log('WS connection closed');
                ws?.close();
                soundboardClient._ws = null;
            };
            ws.onerror = (e) => {
                error(`${e}`);
                soundboardClient._ws = null;
                reject(e);
            };
        });
    },

    async sendCommand(command: string) {
        await soundboardClient._connect();
        log(`Send command: ${command}`);
        soundboardClient._ws?.send(command);
    },

    async play(name: string) {
        await soundboardClient.sendCommand(`play|${name}`);
    },

    async playMultiple(...sounds: PlaySound[]) {
        const command = sounds
            .map(({ name, start, end }) => `${name}:${start}-${end}`)
            .join(',');
        await soundboardClient.sendCommand(`play-multi|${command}`);
    },

    async playMultipleInstruction(instruction: string) {
        await soundboardClient.sendCommand(`play-multi|${instruction}`);
    },

    async stop() {
        await soundboardClient.sendCommand('stop');
    },
};

/** WEB APP */
const MULTI_INSTRUCTION =
    'Play snippets of different sounds back to back.\nUse the format...\n<sound_name>:<start>-<end>,<sound_name>:<start>-<end>,etc\nStart/end are in milliseconds.';

export function init(guildId: string, channelId: string) {
    // init client
    soundboardClient.init(guildId, channelId);

    // soundboard actions
    const onPlayButton = (e: Event) => {
        e.preventDefault();
        let sound = (e.target as Element)?.getAttribute('data-sound');
        if (sound) soundboardClient.play(sound);
    };
    const onStopButton = (e: Event) => {
        e.preventDefault();
        soundboardClient.stop();
    };
    const onMultiButton = (e: Event) => {
        e.preventDefault();
        const instruct = prompt(MULTI_INSTRUCTION);
        if (instruct) soundboardClient.playMultipleInstruction(instruct);
    };

    // bind buttons to functions
    let buttons = document.getElementsByTagName('a');
    for (let i = 0; i < buttons.length; i++) {
        let callbackFunc = onPlayButton;
        switch (buttons[i].id) {
            case 'stop':
                callbackFunc = onStopButton;
                break;
            case 'play-multi':
                callbackFunc = onMultiButton;
                break;
        }
        buttons[i].addEventListener('click', callbackFunc);
    }

    // hook 's' key to stop sounds
    window.addEventListener('keypress', (e) => {
        if (e.key == 's') {
            soundboardClient.stop();
        }
    });
}
