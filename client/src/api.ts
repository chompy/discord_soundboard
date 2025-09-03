import { error, handleNotAuthenticatedError, log } from './utils';

export type User = {
    id: string;
    name: string;
};

export type Guild = {
    id: string;
    name: string;
    icon: string;
};

export type Category = {
    id?: number;
    name: string;
    guildId: string;
    sort?: number;
    created?: Date;
    updated?: Date;
};

export type Sound = {
    id?: number;
    name: string;
    hash: string;
    categoryId: number;
    sort?: number;
    created?: Date;
    updated?: Date;
};

export const api = {
    async _fetch(
        url: string,
        method: string = 'GET',
        data?: object
    ): Promise<any> {
        const resp = await fetch(url, {
            method,
            body: data instanceof ArrayBuffer ? data : JSON.stringify(data),
        });
        if (!resp.ok) {
            const { error: errorMsg } = await resp.json();
            error(errorMsg);
            handleNotAuthenticatedError(errorMsg);
            throw new Error(errorMsg);
        }
        return resp.json();
    },

    async me(): Promise<User> {
        log('Fetch current user');
        const { user } = await api._fetch('/api/me');
        return user;
    },

    async listGuilds(): Promise<Guild[]> {
        log('Fetch user guild list');
        const { guilds } = await api._fetch('/api/list_user_guilds');
        return guilds;
    },

    async listCategories(guildId: string): Promise<Category[]> {
        log(`Fetch categories for guild ${guildId}`);
        const { categories } = await api._fetch(
            '/api/list_guild_categories?guild=' + guildId
        );
        return categories;
    },

    async listSounds(guildId: string): Promise<Sound[]> {
        log(`Fetch sounds for guild ${guildId}`);
        const { sounds } = await api._fetch(
            '/api/list_guild_sounds?guild=' + guildId
        );
        return sounds;
    },

    async uploadSound(data: ArrayBuffer): Promise<string> {
        log(`Upload new sound`);
        const { hash } = await api._fetch('/api/upload_sound', 'POST', data);
        return hash;
    },

    async saveCategory(category: Category): Promise<Category> {
        log(`Save category ${category.name} (${category.id ?? '(new)'})`);
        const { category: output } = await api._fetch(
            '/api/category',
            category.id ? 'PUT' : 'POST',
            category
        );
        return output;
    },

    async sortCategories(categories: Category[]): Promise<void> {
        if (categories.length == 0) return;
        const guildId = categories[0].guildId;
        log(`Sort categories for guild ${guildId}`);
        await api._fetch('/api/sort_guild_categories', 'POST', {
            guildId,
            ids: categories.map((category) => category.id),
        });
    },

    async deleteCategory(category: Category): Promise<void> {
        log(`Delete category ${category.name} (${category.id})`);
        await api._fetch('/api/category', 'DELETE', { id: category.id });
    },

    async saveSound(sound: Sound): Promise<Sound> {
        log(`Save sound ${sound.name} with hash ${sound.hash}`);
        const { sound: output } = await api._fetch(
            '/api/sound',
            sound.id ? 'PUT' : 'POST',
            sound
        );
        return output;
    },

    async sortSounds(sounds: Sound[]): Promise<void> {
        if (sounds.length == 0) return;
        const categoryId = sounds[0].categoryId;
        log(`Sort sounds for category ${categoryId}`);
        await api._fetch('/api/sort_category_sounds', 'POST', {
            categoryId,
            ids: sounds.map((sound) => sound.id),
        });
    },

    async deleteSound(sound: Sound): Promise<void> {
        log(`Delete sound ${sound.name} (${sound.id})`);
        await api._fetch('/api/sound', 'DELETE', { id: sound.id });
    },

    async playSound(sound: Sound): Promise<void> {
        log(`Play sound ${sound.name} (${sound.id})`);
        await api._fetch('/api/play_sound', 'POST', { id: sound.id });
    },

    async stopSounds(guildId: string): Promise<void> {
        log(`Stop all sounds`);
        await api._fetch('/api/stop_sounds', 'POST', { guildId });
    },
};
