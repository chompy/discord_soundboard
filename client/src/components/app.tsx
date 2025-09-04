import '../scss/app.scss';
import React, { useCallback, useEffect, useState } from 'react';
import Button from './button';
import GuildSelect from './guild_select';
import Modal from './modal';
import SoundAdmin from './sound_admin';
import { api, Guild } from '../api';
import { isNotAuthenticatedError, log } from '../utils';
import SoundPlayer from './sound_player';
import useSoundList from '../hooks/sound_list';

export type ModalType = 'admin' | null;

function AppComponent() {
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [activeModal, setActiveModal] = useState<ModalType>(null);
    const [activeGuild, setActiveGuild] = useState<Guild | null>(null);
    const [modalHeight, setModalHeight] = useState(0);
    const soundList = useSoundList(activeGuild?.id);

    const stopSounds = useCallback(() => {
        activeGuild && api.stopSounds(activeGuild.id);
    }, [activeGuild]);

    const onPressStop = useCallback(
        (e: KeyboardEvent) => {
            if (e.key === 's') stopSounds();
        },
        [activeGuild]
    );

    useEffect(() => {
        activeGuild && log(`Set active guild to ${activeGuild.id}`);
        window.addEventListener('keypress', onPressStop);
        return () => {
            window.removeEventListener('keypress', onPressStop);
        };
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
                <h1>Chompy's Discord Soundboard</h1>
            </div>

            <div className="options">
                <Button label="Refresh" onClick={() => soundList.refresh()} />
                <Button label="Stop All [s]" onClick={stopSounds} />
                <Button
                    label="Edit Sounds"
                    onClick={() => setActiveModal('admin')}
                />
            </div>

            {activeGuild && <SoundPlayer soundList={soundList} />}
        </div>
    );
}

function App() {
    return <AppComponent />;
}

export default App;
