import React from 'react';
import { api, Guild } from '../api';
import Select from './select';

export type GuildSelectProperties = {
    onChange?: (guild: Guild | null) => void;
};

function GuildSelect({ onChange }: GuildSelectProperties) {
    const [guilds, setGuilds] = React.useState<Guild[]>([]);
    React.useEffect(() => {
        api.listGuilds().then((guilds) => {
            onChange(guilds[0]);
            setGuilds(guilds);
        });
    }, []);

    return (
        <Select
            options={guilds.map((guild) => guild.name)}
            onChange={(index) => onChange(guilds[index])}
        />
    );
}

export default GuildSelect;
