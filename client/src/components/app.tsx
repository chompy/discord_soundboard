import '../scss/app.scss';
import React from 'react';
import Button from './button';
import GuildSelect from './guild_select';
import Modal from './modal';
import SoundAdmin from './sound_admin';
import { Guild } from '../api';

export type ModalType = 'admin' | null;

function AppComponent() {
    const [activeModal, setActiveModal] = React.useState<ModalType>(null);
    const [activeGuild, setActiveGuild] = React.useState<Guild | null>(null);

    return (
        <div className="app">
            <Modal
                isOpen={activeModal == 'admin'}
                close={() => setActiveModal(null)}
            >
                {activeGuild && <SoundAdmin guildId={activeGuild.id} />}
            </Modal>

            <div className="header">
                <GuildSelect onChange={setActiveGuild} />
                <h1>Chompy's Discord Soundboard</h1>
            </div>

            <div className="options">
                <Button
                    label="Edit Sounds"
                    onClick={() => setActiveModal('admin')}
                />
            </div>

            <p>test</p>
        </div>
    );
}

function App() {
    return <AppComponent />;
}

export default App;
