import { log } from './utils';

export const sound = {
    convert: (soundInput: ArrayBuffer) => {
        log('Convert sound');

        return new Promise<ArrayBuffer>(async (resolve, reject) => {
            const ctx = new AudioContext();

            // decode sound input
            const decodedBuffer = await ctx.decodeAudioData(soundInput);

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
    },
};
