import '../scss/app.scss';
import React, { useCallback, useEffect, useRef, useState } from 'react';
import Button from './button';
import GuildSelect from './guild_select';
import Modal from './modal';
import SoundAdmin from './sound_admin';
import { api, Guild, Sound } from '../api';
import { getSoundKeybindForKey, isNotAuthenticatedError, log } from '../utils';
import SoundPlayer from './sound_player';
import useSoundList from '../hooks/sound_list';

export type ModalType = 'admin' | null;

function AppComponent() {
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [activeModal, setActiveModal] = useState<ModalType>(null);
    const [activeGuild, setActiveGuild] = useState<Guild | null>(null);
    const [modalHeight, setModalHeight] = useState(0);
    const [keyPressEnabled, setKeyPressEnabled] = useState(true);
    const [playerKeyPressEnabled, setPlayerKeyPressEnabled] = useState(true);
    const soundList = useSoundList(activeGuild?.id);
    const filterInputRef = useRef<HTMLInputElement>(null);

    const stopSounds = useCallback(() => {
        activeGuild && api.stopSounds(activeGuild.id);
    }, [activeGuild]);

    const onKeyDown = useCallback((e: KeyboardEvent) => {
        if (!keyPressEnabled) return;
        switch (e.key) {
            case 's':
                // stop all sound playback
                stopSounds();
                break;
            case 'Escape':
                // clear filter
                filterInputRef.current.value = '';
                soundList.sounds.setFilter('');
                break;
            default:
                // play sound from keybind
                const soundId = getSoundKeybindForKey(e.key);
                if (soundId !== null) {
                    const sound = soundList.sounds.get().find((sound) => sound.id === soundId);
                    sound && api.playSound(sound)
                }
                break;
        }
    }, [keyPressEnabled, activeGuild, soundList])

    useEffect(() => {
        window.addEventListener('keydown', onKeyDown)
        return () => {
            window.removeEventListener('keydown', onKeyDown)
        }
    }, [keyPressEnabled, activeGuild, soundList]);

    useEffect(() => {
        activeGuild && log(`Set active guild to ${activeGuild.id}`);
    }, [activeGuild]);

    useEffect(() => {
        api.me()
            .then((user) => {
                log(`Logged in as ${user.name} (${user.id})`);
                setIsLoading(false);
            })
            .catch((error) => {
                if (isNotAuthenticatedError(error)) {
                    window.location.href = '/login';
                    return;
                }
                setError(`${error}`);
            });
    }, []);

    const onFilter = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
        soundList.sounds.setFilter(e.target.value)
    }, [soundList])

    const onKeyBindSoundSelect = (sound: Sound | null) => setKeyPressEnabled(!sound);

    if (error) {
        return <div className="error">{error}</div>;
    }

    if (isLoading) {
        return <div className="loading">Loading...</div>;
    }

    return (
        <div className="app">
            <Modal
                isOpen={activeModal == 'admin'}
                close={() => setActiveModal(null)}
                onResize={setModalHeight}
            >
                {activeGuild && (
                    <SoundAdmin
                        soundList={soundList}
                        guildId={activeGuild.id}
                        height={modalHeight}
                    />
                )}
            </Modal>

            <div className="header">
                <GuildSelect onChange={setActiveGuild} />
                <h1><img src="/static/favicon.png" alt="Chompy" /> Soundboard</h1>
            </div>

            <div className="options">
                <Button label="Refresh" onClick={() => soundList.refresh()} />
                <Button label="Stop All [s]" onClick={stopSounds} />
                <Button
                    label="Edit Sounds"
                    onClick={() => setActiveModal('admin')}
                />
            </div>

            <div className="filter">
                <input 
                    id="filter"
                    type="text"
                    ref={filterInputRef}
                    onFocus={() => { setKeyPressEnabled(false); setPlayerKeyPressEnabled(false); }}
                    onBlur={() => { setKeyPressEnabled(true); setPlayerKeyPressEnabled(true); }}
                    onChange={onFilter}
                    placeholder='Filter'
                />
            </div>

            {activeGuild && <SoundPlayer soundList={soundList} enableKeyBinding={playerKeyPressEnabled} onSelect={onKeyBindSoundSelect} />}
        </div>
    );
}

function App() {
    return <AppComponent />;
}

export default App;
