<template>
    <b-row class="mb-5">
        <b-col>
            <b-card header="Definition">
                <b-form @submit="onSubmit">

                    <b-form-group label-cols="2" id="template_id" label="Definition ID" label-for="template_id_input">

                        <b-input-group>
                            <b-form-input id="template_id_input" type="text" v-model="form.id" placeholder="Definition ID" @change="autoLoadTemplate()"></b-form-input>
                            <b-input-group-append>
                                <b-button variant="success" @click="loadTemplate(false)">Load</b-button>
                            </b-input-group-append>
                        </b-input-group>
                    </b-form-group>

                    <b-form-group label-cols="2" id="template_type" label="Image Type" label-for="template_type_input">
                        <b-form-select id="template_type_input" v-model="form.type"
                                       :options="type_options"></b-form-select>
                    </b-form-group>

                    <b-form-group label-cols="2" id="template_content" label="Graphviz Definition"
                                  label-for="template_content_input">
                        <codemirror v-model="form.def" class="mt-3 adanos-code-textarea" :options="templateOption"></codemirror>
                    </b-form-group>

                    <b-button type="submit" variant="primary" class="mr-2">Submit</b-button>
                </b-form>
            </b-card>
            <b-card header="Preview" v-if="resp !== null" class="mt-3">
                <b-overlay :show="show_overlay" rounded="sm">
                    <b-table striped hover :items="data_table" :fields="data_table_header"></b-table>
                    <b-img :src="$store.getters.serverUrl + resp.preview" style="max-width: 100%;"></b-img>
                </b-overlay>
            </b-card>
        </b-col>
    </b-row>
</template>

<script>
    import axios from 'axios';
    import {codemirror} from 'vue-codemirror-lite';

    export default {
        name: 'Home',
        components: {codemirror},
        data() {
            return {
                form: {
                    id: '',
                    type: 'svg',
                    def: '',
                },
                type_options: ["svg", "svgz", "webp", "png", "bmp", "jpg", "pdf", "gif"],
                resp: null,
                data_table_header: [
                    {key: 'key', label: 'Key'},
                    {key: 'value', label: 'Value'},
                ],
                data_table: [],
                show_overlay: false,
                templateOption: {
                    smartIndent: true,
                    completeSingle: false,
                    lineNumbers: true,
                    lineWrapping: true
                },
            };
        },
        methods: {
            loadMore() {

            },
            onSubmit(evt) {
                evt.preventDefault();
                this.show_overlay = true;

                let req = this.form;
                axios.post('/api/graphviz/definition', req).then((resp) => {
                    this.ToastSuccess('OK');
                    this.resp = resp.data;
                    this.data_table = [];
                    for (let k in this.resp) {
                        let value = this.resp[k];
                        if (k.startsWith('preview')) {
                            value = this.$store.getters.serverUrl.replace(/\/+$/, '') + '/' + value.replace(/^\/+/, '');
                        }

                        this.data_table.push({key: k, value: value});
                    }
                    this.show_overlay = false;
                }).catch(error => {
                    this.show_overlay = false;
                    this.ErrorBox(error)
                });
            },
            autoLoadTemplate() {
                if (this.form.def !== '') {
                    return ;
                }

                this.loadTemplate(true);
            },
            loadTemplate(autoLoad) {
                axios.get('/api/graphviz/definition', {params: {id: this.form.id}}).then((resp) => {
                    this.ToastSuccess('Loaded');
                    this.form.def = resp.data.def;
                }).catch(error => {
                    if (!autoLoad) {
                        this.ToastError(error)
                    }
                });
            },
        },
        mounted() {
            this.loadMore();
        }
    }
</script>

<style scoped>

</style>