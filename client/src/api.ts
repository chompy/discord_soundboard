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

export const isCategory = (value: unknown) =>
    typeof value === 'object' && value && 'guildId' in value;

export type Sound = {
    id?: number;
    name: string;
    hash: string;
    categoryId: number;
    sort?: number;
    created?: Date;
    updated?: Date;
};

export type UserFavorite = {
    userId: number;
    soundId: number;
    created: Date;
}

export type UserKeybind = {
    userId: number;
    soundId: number;
    key: string | null;
    created: Date;
}


export const isSound = (value: unknown) =>
    typeof value === 'object' && value && 'hash' in value;

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

    async listCategoriesAndSounds(
        guildId: string
    ): Promise<[Category[], Sound[]]> {
        log(`Fetch categories and sounds for guild ${guildId}`);
        const { categories, sounds } = await api._fetch(
            '/api/list_guild_categories_and_sounds?guild=' + guildId
        );
        return [categories, sounds];
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
            ids: sounds.flatMap((sound) =>
                sound.categoryId === categoryId ? [sound.id] : []
            ),
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

    async downloadSound(sound: Sound): Promise<ArrayBuffer> {
        log(`Download sound ${sound.name} (${sound.id})`);

        const resp = await fetch('/api/download_sound', {
            method: 'POST',
            body: JSON.stringify({id: sound.id}),
        });
        if (!resp.ok) {
            const { error: errorMsg } = await resp.json();
            error(errorMsg);
            handleNotAuthenticatedError(errorMsg);
            throw new Error(errorMsg);
        }
        return await resp.arrayBuffer()
    },


    async stopSounds(guildId: string): Promise<void> {
        log(`Stop all sounds`);
        await api._fetch('/api/stop_sounds', 'POST', { guildId });
    },

    async listUserFavorites(): Promise<UserFavorite[]> {
        log(`Fetch favorite sounds`);
        const { favorites } = await api._fetch('/api/user_favorite')
        return favorites;
    },

    async addUserFavorite(sound: Sound): Promise<void> {
        log(`Add favorite sound ${sound.name} (${sound.id})`);
        await api._fetch('/api/user_favorite', 'POST', {soundId: sound.id})
    },

    async deleteUserFavorite(sound: Sound): Promise<void> {
        log(`Delete favorite sound ${sound.name} (${sound.id})`);
        await api._fetch('/api/user_favorite', 'DELETE', {soundId: sound.id})
    },
    
    async listUserKeybinds(): Promise<UserKeybind[]> {
        log(`Fetch keybinds`);
        const { keybinds } = await api._fetch('/api/user_keybind')
        return keybinds;
    },

    async addUserKeybind(sound: Sound, key: string): Promise<void> {
        log(`Save keybind ${key} for ${sound.name} (${sound.id})`);
        await api._fetch('/api/user_keybind', 'POST', {soundId: sound.id, key})
    },

    async deleteUserKeybindForSound(sound: Sound): Promise<void> {
        log(`Delete keybind for sound ${sound.name} (${sound.id})`);
        await api._fetch('/api/user_keybind', 'DELETE', {soundId: sound.id})
    },

    async deleteUserKeybindForKey(key: string): Promise<void> {
        log(`Delete keybind for key ${key}`);
        await api._fetch('/api/user_keybind', 'DELETE', {key})
    } 
};
