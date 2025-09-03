import React, { useEffect } from 'react';
import { api, Guild } from '../api';
import Select from './select';

export type GuildSelectProperties = {
    onChange?: (guild: Guild | null) => void;
};

function GuildSelect({ onChange }: GuildSelectProperties) {
    const [guilds, setGuilds] = React.useState<Guild[]>([]);
    useEffect(() => {
        api.listGuilds().then((guilds) => {
            setGuilds(guilds);
            onChange?.(guilds[0]);
        });
    }, []);

    return (
        <Select
            options={guilds.map((guild) => guild.name)}
            onChange={(index) => onChange?.(guilds[index])}
        />
    );
}

export default GuildSelect;
