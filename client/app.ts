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

    async upload(buffer: ArrayBuffer) {
        log('Upload sound.');
        await fetch(
            `/upload?guild=${soundboardClient._guildId}&channel=${soundboardClient._channelId}`,
            {
                method: 'POST',
                body: buffer,
            }
        );
    },
};

/** SOUND CONVERTER */
const convertSound = (soundInput: ArrayBuffer) => {
    log('Convert sound');

    return new Promise<ArrayBuffer>(async (resolve, reject) => {
        // decode sound input
        const decodedBuffer = await new AudioContext().decodeAudioData(
            soundInput
        );

        // setup opus encoder
        const encodedChunks: ArrayBuffer[] = [];
        const encoder = new AudioEncoder({
            output: (chunk) => {
                // get chunk bytes and push to array of chunks
                const chunkBuffer = new ArrayBuffer(chunk.byteLength);
                chunk.copyTo(chunkBuffer);
                encodedChunks.push(chunkBuffer);
            },
            error: reject,
        });
        encoder.configure({
            codec: 'opus',
            bitrate: 128000,
            sampleRate: decodedBuffer.sampleRate,
            numberOfChannels: decodedBuffer.numberOfChannels,
        });

        // interleave two channel (stero) audio
        if (decodedBuffer.numberOfChannels == 0) {
            reject(new Error('cannot read sound with zero channels'));
            return;
        }
        const leftChannelData = decodedBuffer.getChannelData(0);
        const rightChannelData =
            decodedBuffer.numberOfChannels >= 2
                ? decodedBuffer.getChannelData(1)
                : null;
        const interleavedData = new Float32Array(
            leftChannelData.length +
                (rightChannelData ? rightChannelData.length : 0)
        );
        for (
            let src = 0, dst = 0;
            src < leftChannelData.length;
            src++, dst += rightChannelData ? 2 : 1
        ) {
            interleavedData[dst] = leftChannelData[src];
            if (rightChannelData) {
                interleavedData[dst + 1] = rightChannelData[src];
            }
        }

        const audioData = new AudioData({
            format: 'f32',
            numberOfChannels: decodedBuffer.numberOfChannels,
            numberOfFrames: decodedBuffer.length,
            sampleRate: decodedBuffer.sampleRate,
            timestamp: 0,
            data: interleavedData,
        });

        // start encoder
        encoder.encode(audioData);
        await encoder.flush();

        // add chunks to array
        resolve(
            await new Blob(
                encodedChunks.flatMap((c) => [
                    Uint16Array.from([c.byteLength]).buffer,
                    c,
                ])
            ).arrayBuffer()
        );
    });
};

/** WEB APP */
const MULTI_INSTRUCTION =
    'Play snippets of different sounds back to back.\nUse the format...\n<sound_name>:<start>-<end>,<sound_name>:<start>-<end>,etc\nStart/end are in milliseconds.';

export function init(guildId: string, channelId: string) {
    // init client
    soundboardClient.init(guildId, channelId);

    // handle file upload
    const uploadElement = document.getElementById('upload') as HTMLInputElement;
    uploadElement.addEventListener('change', async (e: Event) => {
        e.preventDefault();
        if (
            e.target &&
            'files' in e.target &&
            e.target.files instanceof FileList &&
            e.target.files.length > 0
        ) {
            const data = await convertSound(
                await e.target.files[0].arrayBuffer()
            );
            await soundboardClient.upload(data);
            uploadElement.value = '';
        }
    });

    // soundboard actions
    const onPlayButton = (e: Event) => {
        e.preventDefault();
        let sound = (e.target as Element)?.getAttribute('data-sound');
        if (sound) soundboardClient.play(sound);
    };
    const onPreviewButton = (e: Event) => {
        e.preventDefault();
        let sound = (e.target as Element)?.getAttribute('data-sound');
        if (!sound) {
            console.error('data-sound attribute missing');
            return;
        }
        new Audio('/download?sound=' + sound).play();
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
    const onFileButton = (e: Event) => {
        e.preventDefault();
        uploadElement.click();
    };

    // bind buttons to functions
    let buttons = document.getElementsByTagName('a');
    for (let i = 0; i < buttons.length; i++) {
        let callbackFunc: ((e: Event) => void) | null = null;
        switch (buttons[i].getAttribute('data-action')) {
            case 'play':
                callbackFunc = onPlayButton;
                break;
            case 'stop':
                callbackFunc = onStopButton;
                break;
            case 'play-multi':
                callbackFunc = onMultiButton;
                break;
            case 'preview':
                callbackFunc = onPreviewButton;
                break;
            case 'play-file':
                callbackFunc = onFileButton;
                break;
        }
        if (callbackFunc) buttons[i].addEventListener('click', callbackFunc);
    }

    // hook 's' key to stop sounds
    window.addEventListener('keypress', (e) => {
        if (e.key == 's') {
            soundboardClient.stop();
        }
    });
}
